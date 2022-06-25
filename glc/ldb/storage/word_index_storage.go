/**
 * 关键词反向索引存储器
 * 1）获取存储对象线程安全，带缓存无则创建有则直取，空闲超时自动关闭leveldb，再次获取时自动打开
 * 2）单线程调用设计，由日志存储器内部控制安全的调用，其他地方调用可能会有问题
 */
package storage

import (
	"glc/cmn"
	"glc/ldb/conf"
	"log"
	"math"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

// 关键词索引存储结构体
type WordIndexStorage struct {
	storeName    string      // 存储目录
	subPath      string      // 存储目录下的相对路径（存放数据）
	word         string      // 索引关键词
	leveldb      *leveldb.DB // leveldb
	currentCount uint64      // 当前件数
	lastTime     int64       // 最后一次访问时间
	closing      bool        // 是否关闭中状态
	mu           sync.Mutex  // 锁
}

var idxMu sync.Mutex
var mapWordIndexStorage map[string](*WordIndexStorage)

func init() {
	mapWordIndexStorage = make(map[string](*WordIndexStorage))
}

func getWidxStorage(cacheName string) *WordIndexStorage {
	cacheStore := mapWordIndexStorage[cacheName]
	if cacheStore != nil && !cacheStore.IsClose() {
		return cacheStore // 缓存中未关闭的存储对象
	}
	return nil
}

// 获取存储对象，线程安全（带缓存无则创建有则直取）
func NewWordIndexStorage(storeName string, word string) *WordIndexStorage { // 存储器，文档，自定义对象

	// 缓存有则取用
	subPath := "inverted" + cmn.PathSeparator() + cmn.HashAndMod(word, 100) + cmn.PathSeparator() + "k_" + cmn.HashAndMod(word, math.MaxUint32)
	cacheName := storeName + cmn.PathSeparator() + subPath
	cacheStore := getWidxStorage(cacheName)
	if cacheStore != nil && !cacheStore.IsClose() {
		return cacheStore
	}

	// 缓存无则锁后创建返回并存缓存
	idxMu.Lock()                           // 上锁
	cacheStore = getWidxStorage(cacheName) // 再次尝试取用缓存中存储器
	if cacheStore != nil && !cacheStore.IsClose() {
		idxMu.Unlock() // 解锁
		return cacheStore
	}

	store := new(WordIndexStorage)
	store.storeName = storeName
	store.subPath = subPath
	store.word = word
	store.closing = false
	store.lastTime = time.Now().Unix()

	dbPath := conf.GetStorageRoot() + cmn.PathSeparator() + cacheName
	db, err := leveldb.OpenFile(dbPath, nil) // 打开（在指定子目录中存放数据）
	if err != nil {
		log.Println("打开WordIndexStorage失败：", dbPath)
		panic(err)
	}
	store.leveldb = db
	store.currentCount = store.loadTotalCount() // 读取总件数
	mapWordIndexStorage[cacheName] = store      // 缓存起来

	// 逐秒判断，若闲置超时则自动关闭
	go autoCloseWordIndexStorageWhenMaxIdle(store)

	idxMu.Unlock() // 解锁
	log.Println("打开WordIndexStorage：", cacheName)
	return store
}

func autoCloseWordIndexStorageWhenMaxIdle(store *WordIndexStorage) {
	if conf.GetMaxIdleTime() > 0 {
		ticker := time.NewTicker(time.Second)
		for {
			<-ticker.C
			if time.Now().Unix()-store.lastTime > int64(conf.GetMaxIdleTime()) {
				store.Close()
				ticker.Stop()
				break
			}
		}
	}
}

// 日志ID添加到索引
func (s *WordIndexStorage) Add(id uint64) error {

	// 加索引
	s.lastTime = time.Now().Unix()
	s.currentCount++ // ID递增
	err := s.leveldb.Put(cmn.Uint64ToBytes(s.currentCount), cmn.Uint64ToBytes(id), nil)
	if err != nil {
		log.Println("保存索引失败", err)
		return err
	}

	// 保存建好的索引数
	s.leveldb.Put(cmn.Uint64ToBytes(0), cmn.Uint64ToBytes(s.currentCount), nil)
	if err != nil {
		log.Println("保存索引件数失败", err)
		return err // 忽略事务问题，可下回重建
	}
	log.Println("创建日志索引：", id, "，关键词：", s.word)
	return nil
}

// 通过索引ID取日志ID（返回0表示有问题）
func (s *WordIndexStorage) Get(id uint64) uint64 {
	if s.closing {
		return 0
	}
	s.lastTime = time.Now().Unix()
	b, err := s.leveldb.Get(cmn.Uint64ToBytes(id), nil)
	if err != nil {
		return 0
	}
	return cmn.BytesToUint64(b)
}

// 关闭Storage
func (s *WordIndexStorage) Close() {
	if s.closing {
		return
	}

	s.mu.Lock() // 锁
	if s.closing {
		s.mu.Unlock() // 解锁
		return
	}

	s.closing = true
	s.leveldb.Close()                      // 走到这里时没有db操作了，可以关闭
	mapWordIndexStorage[s.storeName] = nil // 设空，下回GetStorage时自动再创建
	s.mu.Unlock()                          // 解锁

	log.Println("关闭WordIndexStorage：", s.storeName+cmn.PathSeparator()+s.subPath)
}

func (s *WordIndexStorage) loadTotalCount() uint64 {
	bytes, err := s.leveldb.Get(cmn.Uint64ToBytes(0), nil)
	if err != nil || bytes == nil {
		return 0
	}
	return cmn.BytesToUint64(bytes)
}

// 总件数
func (s *WordIndexStorage) TotalCount() uint64 {
	return s.currentCount
}

// 存储目录名
func (s *WordIndexStorage) StoreName() string {
	return s.storeName
}

// 是否关闭中状态
func (s *WordIndexStorage) IsClose() bool {
	return s.closing
}