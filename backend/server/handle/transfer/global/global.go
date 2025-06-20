package global

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/sftp"

	"golang.org/x/crypto/ssh"
)

type SSHConnection struct {
	Client *ssh.Client
	UsedAt time.Time
}

type SSHConnectionPool struct {
	sync.Mutex
	Connections map[string]*SSHConnection // key可以是server IP或者标识符
	Capacity    int                       // 每个服务器的最大连接数
	Timeout     time.Duration             // 连接的最大空闲时间
}

// 定义一个具体类型来实现FileTransferService接口
type FileTransferServiceImpl struct {
	Pool *SSHConnectionPool
}

type FileTransferService interface {
	CreateTransferTask(srcServer, destServer, srcPath, destPath string) (string, error)
	GetTransferStatus(taskID string) (string, error)
}

var Pool *SSHConnectionPool // 全局连接池

var FTS *FileTransferServiceImpl // 全局文件传输服务

// 添加连接到连接池中
func (p *SSHConnectionPool) Add(server string, client *ssh.Client) {
	p.Lock()
	defer p.Unlock()

	if !isConnectionValid(client) {
		fmt.Println("Add：无效的连接")
		return
	}

	// 如果已有连接，先关闭旧的，然后更新连接
	if oldConn, exists := p.Connections[server]; exists {
		if oldConn.Client == client { // 新旧连接相同
			// 添加/更新连接
			p.Connections[server] = &SSHConnection{
				Client: client,
				UsedAt: time.Now(),
			}
			return
		}
		oldConn.Client.Close()        // 关闭旧连接
		delete(p.Connections, server) // 删除旧连接
	}

	// 没有旧连接，或删除了旧连接，则判断容量
	if len(p.Connections) >= p.Capacity {
		client.Close() // 超过容量则关闭连接
		return
	}

	// 添加新连接
	p.Connections[server] = &SSHConnection{
		Client: client,
		UsedAt: time.Now(),
	}
}

// 从连接池中获取连接
func (p *SSHConnectionPool) Get(server string) (*ssh.Client, error) {
	p.Lock()
	defer p.Unlock()

	if conn, exists := p.Connections[server]; exists {
		if time.Since((*conn).UsedAt) > p.Timeout || !isConnectionValid((*conn).Client) {
			conn.Client.Close() // 连接超时，关闭并删除
			delete(p.Connections, server)
			return nil, errors.New("Get：连接已过期或无效")
		} else {
			(*conn).UsedAt = time.Now() // 更新使用时间
			return (*conn).Client, nil
		}
	}
	// 如果不存在有效连接，返回错误
	return nil, errors.New("Get：没有可用连接")
}

// 放回/更新连接到连接池中
func (p *SSHConnectionPool) Put(server string, client *ssh.Client) {
	p.Lock()
	defer p.Unlock()

	// 判断新连接是否有效
	if !isConnectionValid(client) {
		fmt.Println("Put：无效的连接")
		if client == p.Connections[server].Client { // 新旧连接相同
			client.Close()                // 关闭旧连接
			delete(p.Connections, server) // 删除旧连接
		}
		return
	}

	// 如果已有连接，先判断是否有效
	if oldConn, exists := p.Connections[server]; exists {
		if isConnectionValid(oldConn.Client) {
			oldConn.UsedAt = time.Now() // 更新使用时间
			// 不client.Close()，因为这个client与oldConn.Client是同一个
			// 只是更新了oldConn.UsedAt，而不是关闭oldConn.Client
			p.Connections[server] = oldConn
			return
		} else {
			_ = oldConn.Client.Close()    // 关闭旧连接
			delete(p.Connections, server) // 删除旧连接
		}
	}

	// 没有旧连接或旧连接已失效，则判断容量，准备添加新连接
	if len(p.Connections) >= p.Capacity {
		client.Close() // 超过容量则关闭连接
		return
	}

	// 添加新连接
	p.Connections[server] = &SSHConnection{
		Client: client,
		UsedAt: time.Now(),
	}
}

func isConnectionValid(client *ssh.Client) bool {
	session, err := client.NewSession()
	if err != nil {
		return false
	}
	defer session.Close()

	// _, err = session.StdoutPipe() // 尝试获取会话输出流，检查连接是否有效
	_, err = session.CombinedOutput("whoami") // 检查连接是否有效
	return err == nil
}

// 定期清理连接池
func (p *SSHConnectionPool) Cleanup(stopChan chan struct{}) {
	ticker := time.NewTicker(p.Timeout / 2) // 每半超时时间检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.Lock()
			now := time.Now()
			for server, conn := range p.Connections {
				if now.Sub(conn.UsedAt) > p.Timeout {
					conn.Client.Close() // 连接超时，关闭并删除
					delete(p.Connections, server)
				}
			}
			p.Unlock()
		case <-stopChan:
			return // 停止清理
		}
	}
}

// 创建普通传输任务：客户端上传文件给指定服务器
func (fts *FileTransferServiceImpl) CreateCommonUploadTask(file *multipart.FileHeader, server, path string) (string, error) {
	// 获取连接（不放回，因为传输过程中需要保持连接）
	client, err := fts.Pool.Get(server)
	if err != nil {
		fmt.Printf("获取连接失败: %v\n", err)
		return "", err
	}
	defer fts.Pool.Put(server, client)

	// 创建SFTP客户端
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		fmt.Printf("创建SFTP客户端失败: %v\n", err)
		fts.Pool.Put(server, client) // 放回连接
		return "", err
	}
	defer sftpClient.Close()

	// 实际传输逻辑
	srcFile, err := file.Open()
	if err != nil {
		fmt.Printf("打开文件失败: %v", err)
		return "", err
	}
	defer srcFile.Close()

	destFile, err := sftpClient.Create(path) // 创建远程文件
	if err != nil {
		fmt.Printf("创建远程文件失败: %v", err)
		return "", err
	}
	defer destFile.Close()

	// 复制文件内容
	if _, err := io.Copy(destFile, srcFile); err != nil {
		fmt.Printf("文件复制失败: %v", err)
		return "", err
	}

	// 确保文件权限正确
	if err := sftpClient.Chmod(path, 0644); err != nil { // 假设目标文件需要0644权限
		fmt.Printf("文件权限设置失败: %v", err)
		return "", err
	}

	// 生成任务ID
	taskID := uuid.New().String()

	return taskID, nil
}

// 创建普通传输任务：客户端下载文件给指定服务器
func (fts *FileTransferServiceImpl) CreateCommonDownloadTask(server, path string) (*sftp.Client, string, error) {
	// 获取连接
	client, err := fts.Pool.Get(server)
	if err != nil {
		return nil, "", err
	}
	defer fts.Pool.Put(server, client)
	// 创建SFTP客户端
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		fts.Pool.Put(server, client) // 放回连接
		return nil, "", err
	}
	// defer sftpClient.Close() // 不关闭，后面需要使用

	// 生成任务ID
	taskID := uuid.New().String()

	return sftpClient, taskID, nil
}

// 创建两个服务器间的传输任务
func (fts *FileTransferServiceImpl) CreateTransferBetween2STask(srcServer, srcPath, destServer, destPath string) (string, error) {
	// 获取连接（不放回，因为传输过程中需要保持连接）
	srcClient, err := fts.Pool.Get(srcServer)
	if err != nil {
		return "", err
	}
	defer fts.Pool.Put(srcServer, srcClient)

	destClient, err := fts.Pool.Get(destServer)
	if err != nil {
		return "", err
	}
	defer fts.Pool.Put(destServer, destClient)

	// 创建SFTP客户端
	srcSftp, err := sftp.NewClient(srcClient)
	if err != nil {
		return "", err
	}
	defer srcSftp.Close()

	destSftp, err := sftp.NewClient(destClient)
	if err != nil {
		return "", err
	}
	defer destSftp.Close()

	// 实际传输逻辑
	srcFile, err := srcSftp.Open(srcPath)
	if err != nil {
		log.Printf("打开源文件失败: %v", err)
		return "", err
	}
	defer srcFile.Close()

	destFile, err := destSftp.Create(destPath)
	if err != nil {
		log.Printf("创建目标文件失败: %v", err)
		return "", err
	}
	defer destFile.Close()

	// 复制文件内容
	if _, err := io.Copy(destFile, srcFile); err != nil {
		log.Printf("文件复制失败: %v", err)
		return "", err
	}

	// 确保文件权限正确
	if err := destSftp.Chmod(destPath, 0644); err != nil { // 假设目标文件需要0644权限
		log.Printf("文件权限设置失败: %v", err)
		return "", err
	}

	// 生成任务ID
	taskID := uuid.New().String()

	return taskID, nil
}

func (fts *FileTransferServiceImpl) GetTransferStatus(taskID string) (string, error) {
	// 实现获取任务状态的逻辑
	return "", nil
}