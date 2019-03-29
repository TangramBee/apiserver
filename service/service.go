package service

import (
	"apiserver/model"
	"apiserver/util"
	"fmt"
	"sync"
)

func ListUser(username string, offset, limit int) ([]*model.UserInfo, uint64, error) {
	infos := make([]*model.UserInfo, 0)
	users, count, err := model.ListUser(username, offset, limit)
	if err != nil {
		return nil, count, err
	}

	ids := []uint64{}
	for _, user := range users {
		ids = append(ids, user.Id)
	}

	wg := sync.WaitGroup{}
	userList := model.UserList{
		Lock:  new(sync.Mutex),
		IdMap: make(map[uint64]*model.UserInfo, len(users)),
	}

	errChan := make(chan error, 1)
	finished := make(chan bool, 1)

	for _, u := range users {
		wg.Add(1)
		go func(u *model.UserModel) {
			defer wg.Done()

			shortId, err := util.GenShortId()
			if err != nil {
				errChan <- err // 当出错则写入err通道，挂起这个go协程，继续for下一个go协程
				return
			}

			// 加锁，保持数据一致性
			userList.Lock.Lock()
			defer userList.Lock.Unlock()
			userList.IdMap[u.Id] = &model.UserInfo{
				Id:        u.Id,
				Username:  u.Username,
				SayHello:  fmt.Sprintf("Hello %s", shortId),
				Password:  u.Password,
				CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt: u.UpdatedAt.Format("2006-01-02 15:04:05"),
			}
		}(u)
	}

	// 开启另外的协程，等待数据获取完，close通道chan
	go func() {
		wg.Wait() // 阻塞，直到wg为0
		close(finished)
	}()

	// 请求完数据，则从通道获取数据。
	// 注意：当close channel时，
	// 读取channel的一方会从channel中读取到value,false，
	// 此时的value一般情况下为nil。
	// 该代码也可以用来通知当不使用channel时，关闭channel的情况。
	select {
	case <-finished:
	case err := <-errChan:
		return nil, count, err
	}

	for _, id := range ids {
		infos = append(infos, userList.IdMap[id])
	}

	return infos, count, nil

}
