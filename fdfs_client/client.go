package fdfs_client

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	/*"strconv"
	"strings"*/

	"github.com/sirupsen/logrus"
)

var (
	logger                                          = logrus.New()
	storagePoolChan      chan *storagePool          = make(chan *storagePool, 1)
	storagePoolMap       map[string]*ConnectionPool = make(map[string]*ConnectionPool)
	fetchStoragePoolChan chan interface{}           = make(chan interface{}, 1)
	quit                 chan bool
)

type FdfsClient struct {
	tracker     *Tracker
	trackerPool *ConnectionPool
	timeout     int
}

type Tracker struct {
	HostList []string
	Ports    []int
}
type storagePool struct {
	storagePoolKey string
	hosts          []string
	ports          []int
	minConns       int
	maxConns       int
}

/*func (storagePool storagePool) Print() {
	logger.Info("storagePoolKey:" + storagePool.storagePoolKey)
	logger.Info("hosts" + storagePool.hosts[0])
	logger.Info("ports" + storagePool.ports[0])
}*/

func init() {
	logger.Formatter = new(logrus.TextFormatter)
	//logrus.SetFormatter(&logrus.TextFormatter{})
	//logrus.SetOutput(os.Stdout)
	//logrus.SetLevel(logrus.DebugLevel)
	logger.Level = logrus.DebugLevel
	logger.Out = os.Stdout
	/*logger.WithFields(logrus.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")*/
	//logger.Info("A group of walrus emerges from the ocean")
	runtime.GOMAXPROCS(runtime.NumCPU())
	go func() {
		// start a loop
		for {
			select {
			case spd := <-storagePoolChan:
				if sp, ok := storagePoolMap[spd.storagePoolKey]; ok {
					logger.Debug("storagePool already exist")
					fetchStoragePoolChan <- sp
				} else {
					var (
						sp  *ConnectionPool
						err error
					)
					logger.Debug("starting a new storagePool")
					sp, err = NewConnectionPool(spd.hosts, spd.ports, spd.minConns, spd.maxConns)
					//defer sp.Close()
					if err != nil {
						fetchStoragePoolChan <- err
					} else {
						storagePoolMap[spd.storagePoolKey] = sp
						fetchStoragePoolChan <- sp
					}
				}
			case <-quit:
				break
			}
		}
	}()
}
func getTrackerConf(ConfPath string) (*Tracker, error) {
	Config := &Config{}
	Config, err := getConf(ConfPath)
	if err != nil {
		return nil, err
	}
	tracer := &Tracker{
		HostList: Config.TrackerIp,
		Ports:    Config.TrackerPort,
	}
	//logger.Debugf("tracer.HostList:%s", tracer.HostList)
	//logger.Debugf("tracer.Port:%d", tracer.Port)
	return tracer, nil
}

func NewFdfsClient(confPath string) (*FdfsClient, error) {
	tracker, err := getTrackerConf(confPath)
	if err != nil {
		return nil, err
	}

	trackerPool, err := NewConnectionPool(tracker.HostList, tracker.Ports, MINCONN, MAXCONN)
	if err != nil {
		return nil, err
	}

	return &FdfsClient{tracker: tracker, trackerPool: trackerPool}, nil
}

func NewFdfsClientByTracker(tracker *Tracker) (*FdfsClient, error) {
	trackerPool, err := NewConnectionPool(tracker.HostList, tracker.Ports, MINCONN, MAXCONN)
	if err != nil {
		return nil, err
	}

	return &FdfsClient{tracker: tracker, trackerPool: trackerPool}, nil
}
func ColseFdfsClient() {
	quit <- true
}

func (this *FdfsClient) UploadByFilename(filename string) (*UploadFileResponse, error) {
	if err := fdfsCheckFile(filename); err != nil {
		logger.Error("fdfsCheckFile error" + err.Error())
		return nil, errors.New(err.Error() + "(uploading)")
	}

	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageStorWithoutGroup()
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	return store.storageUploadByFilename(tc, storeServ, filename)
}

func (this *FdfsClient) UploadByBuffer(filebuffer []byte, fileExtName string) (*UploadFileResponse, error) {
	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageStorWithoutGroup()
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	return store.storageUploadByBuffer(tc, storeServ, filebuffer, fileExtName)
}

func (this *FdfsClient) UploadSlaveByFilename(filename, remoteFileId, prefixName string) (*UploadFileResponse, error) {
	if err := fdfsCheckFile(filename); err != nil {
		return nil, errors.New(err.Error() + "(uploading)")
	}

	tmp, err := splitRemoteFileId(remoteFileId)
	if err != nil || len(tmp) != 2 {
		return nil, err
	}
	groupName := tmp[0]
	remoteFilename := tmp[1]

	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageStorWithGroup(groupName)
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	return store.storageUploadSlaveByFilename(tc, storeServ, filename, prefixName, remoteFilename)
}

func (this *FdfsClient) UploadSlaveByBuffer(filebuffer []byte, remoteFileId, fileExtName string) (*UploadFileResponse, error) {
	tmp, err := splitRemoteFileId(remoteFileId)
	if err != nil || len(tmp) != 2 {
		return nil, err
	}
	groupName := tmp[0]
	remoteFilename := tmp[1]

	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageStorWithGroup(groupName)
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	return store.storageUploadSlaveByBuffer(tc, storeServ, filebuffer, remoteFilename, fileExtName)
}

func (this *FdfsClient) UploadAppenderByFilename(filename string) (*UploadFileResponse, error) {
	if err := fdfsCheckFile(filename); err != nil {
		return nil, errors.New(err.Error() + "(uploading)")
	}

	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageStorWithoutGroup()
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	return store.storageUploadAppenderByFilename(tc, storeServ, filename)
}

func (this *FdfsClient) UploadAppenderByBuffer(filebuffer []byte, fileExtName string) (*UploadFileResponse, error) {
	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageStorWithoutGroup()
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	return store.storageUploadAppenderByBuffer(tc, storeServ, filebuffer, fileExtName)
}

func (this *FdfsClient) DeleteFile(remoteFileId string) (*DeleteFileResponse, error) {
	tmp, err := splitRemoteFileId(remoteFileId)
	if err != nil || len(tmp) != 2 {
		return nil, err
	}
	groupName := tmp[0]
	remoteFilename := tmp[1]

	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageUpdate(groupName, remoteFilename)
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	return store.storageDeleteFile(tc, storeServ, remoteFilename)
}

func (this *FdfsClient) DownloadToFile(localFilename string, remoteFileId string, offset int64, downloadSize int64) (*DownloadFileResponse, error) {
	tmp, err := splitRemoteFileId(remoteFileId)
	if err != nil || len(tmp) != 2 {
		return nil, err
	}
	groupName := tmp[0]
	remoteFilename := tmp[1]

	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageFetch(groupName, remoteFilename)
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	return store.storageDownloadToFile(tc, storeServ, localFilename, offset, downloadSize, remoteFilename)
}
func (this *FdfsClient) QueryFileInfo(groupName string, remoteFileName string) (*fileInfo, error) {
	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageFetch(groupName, remoteFileName)
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}
	return store.storageQueryFileInfo(groupName, remoteFileName)
}
func (this *FdfsClient) DownloadToBuffer(remoteFileId string, offset int64, downloadSize int64) (*DownloadFileResponse, error) {
	tmp, err := splitRemoteFileId(remoteFileId)
	if err != nil || len(tmp) != 2 {
		return nil, err
	}
	groupName := tmp[0]
	remoteFilename := tmp[1]

	tc := &TrackerClient{this.trackerPool}
	storeServ, err := tc.trackerQueryStorageFetch(groupName, remoteFilename)
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	store := &StorageClient{storagePool}

	var fileBuffer []byte
	return store.storageDownloadToBuffer(tc, storeServ, fileBuffer, offset, downloadSize, remoteFilename)
}
func (this *FdfsClient) TruncAppenderByFilename(remoteFileId string, truncatedFileSize int64) (*DeleteFileResponse, error) {
	tmp, err := splitRemoteFileId(remoteFileId)
	if err != nil || len(tmp) != 2 {
		return nil, err
	}
	groupName := tmp[0]
	remoteFilename := tmp[1]

	tc := &TrackerClient{this.trackerPool}

	storeServ, err := tc.trackerQueryStorageUpdate(groupName, remoteFilename)
	if err != nil {
		return nil, err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	if err != nil {
		return nil, err
	}

	store := &StorageClient{storagePool}

	return store.storageTruncateFile(tc, storeServ, remoteFilename, truncatedFileSize)
}
func (this *FdfsClient) AppendByFileName(localFileName string, groupName string, remoteFileName string) error {
	tc := &TrackerClient{this.trackerPool}

	storeServ, err := tc.trackerQueryStorageUpdate(groupName, remoteFileName)
	if err != nil {
		return err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	if err != nil {
		return err
	}

	store := &StorageClient{storagePool}
	return store.storageAppendByfileName(tc, storeServ, localFileName, groupName, remoteFileName)
}
func (this *FdfsClient) ModifyByFileName(localFileName string, offset int64, groupName string, remoteFileName string) error {
	tc := &TrackerClient{this.trackerPool}

	storeServ, err := tc.trackerQueryStorageUpdate(groupName, remoteFileName)
	if err != nil {
		return err
	}

	storagePool, err := this.getStoragePool(storeServ.ipAddr, storeServ.port)
	if err != nil {
		return err
	}

	store := &StorageClient{storagePool}
	return store.storageModifyByfileName(tc, storeServ, localFileName, offset, groupName, remoteFileName)
}

func (this *FdfsClient) getStoragePool(ipAddr string, port int) (*ConnectionPool, error) {
	hosts := []string{ipAddr}
	ports := []int{port}
	var (
		storagePoolKey string = fmt.Sprintf("%s-%d", hosts[0], ports[0])
		result         interface{}
		err            error
		ok             bool
	)

	spd := &storagePool{
		storagePoolKey: storagePoolKey,
		hosts:          hosts,
		ports:          ports,
		minConns:       MINCONN,
		maxConns:       MAXCONN,
	}
	storagePoolChan <- spd
	for {
		select {
		case result = <-fetchStoragePoolChan:
			var storagePool *ConnectionPool
			if err, ok = result.(error); ok {
				logger.Error("failed to open connection pool" + err.Error())
				return nil, err
			} else if storagePool, ok = result.(*ConnectionPool); ok {
				return storagePool, nil
			} else {
				Err := errors.New("none operatoin on storagePool yet")
				logger.Error(Err.Error())
				return nil, Err
			}
		}
	}
}
