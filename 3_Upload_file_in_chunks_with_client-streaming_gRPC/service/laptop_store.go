package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"log"
	"pcbook/pb"
	"sync"
)

// ErrAlreadyExists is returned when a record with the same ID already exists in the store
var ErrAlreadyExists = errors.New("record already exists")

// LaptopStore is an interface to store laptop
type LaptopStore interface {
	// Save saves the laptop to the store
	Save(laptop *pb.Laptop) error
	// Find finds a laptop by ID
	Find(id string) (*pb.Laptop, error)
	// Search searches for laptop with filter, returns one by one via the found funtion
	Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error ) error
}

// InMemoryLaptopStore stores laptop in memory
type InMemoryLaptopStore struct {
	mutex sync.RWMutex
	data  map[string]*pb.Laptop
}

func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	// 如果内存中已经存在此笔记本了，返回已存在错误
	if store.data[laptop.Id] != nil {
		return ErrAlreadyExists
	}

	// to be safe, deep copy
	other, err := deepCopy(laptop)
	if err != nil {
		return err
	}

	store.data[other.Id] = other
	return nil
}

func  (store *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	laptop := store.data[id]
	if laptop == nil {
		return nil, nil
	}

	// deep copy
	return deepCopy(laptop)
}

func (store *InMemoryLaptopStore) Search(
	ctx context.Context,
	filter *pb.Filter,
	found func(laptop *pb.Laptop) error,
) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	for _, laptop := range store.data {

		// 模拟超时：假如在这里进行一些繁忙的处理
		// time.Sleep(time.Second)
		// log.Print("check laptop id: ", laptop.GetId())

		if ctx.Err() == context.Canceled || ctx.Err() == context.DeadlineExceeded {
			log.Print("context is canceled")
			return errors.New("context is canceled")
		}

		if isQualified(filter, laptop) {
			other, err := deepCopy(laptop)
			if err != nil {
				return err
			}
			// 调用found()将其发送给调用方
			err  = found(other)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() {
		return false
	}
	if laptop.GetCpu().GetNumberCores() < filter.GetMinCpuCores() {
		return false
	}
	if laptop.GetCpu().GetMinGhz() < filter.GetMinCpuGhz() {
		return false
	}
	// 为了比较它们，将内存转换为最小单位（bit）
	if toBit(laptop.GetRam()) < toBit(filter.GetMinRam()) {
		return false
	}
	return true
}

func toBit(memory *pb.Memory) uint64 {
	value := memory.GetValue()

	switch memory.GetUnit() {
	case pb.Memory_BIT:
		return value
	case pb.Memory_BYTE:
		return value << 3 	// 8 = 2^3
	case pb.Memory_KILOBYTE:
		return value << 13	// 8 * 1024 = 2^3 + 2^10 = 2^13
	case pb.Memory_MEGABATE:
		return value << 23
	case pb.Memory_GIGABYTE:
		return value << 33
	case pb.Memory_TERABYTE:
		return value << 43
	default:
		return 0
	}
}

func deepCopy(laptop *pb.Laptop) (*pb.Laptop, error) {
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data: %w", err)
	}
	return other, nil
}

// DBLaptopStore stores laptop in db
// type DBLaptopStore struct {}
