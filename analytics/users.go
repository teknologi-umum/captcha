package analytics

import (
	"context"

	"github.com/go-redis/redis/v8"
)

func (d *Dependency) GetAllUserID(ctx context.Context) ([]string, error) {
	r, err := d.Redis.SMembers(ctx, "analytics:users").Result()
	if err != nil {
		return []string{}, err
	}

	return r, nil
}

func (d *Dependency) GetAllUserMap(ctx context.Context) ([]UserMap, error) {
	ids, err := d.GetAllUserID(ctx)
	if err != nil {
		return []UserMap{}, err
	}

	var users []UserMap

	tx := d.Redis.TxPipeline()
	defer tx.Close()

	var userCmd = make(map[string]*redis.StringStringMapCmd, len(ids))

	for _, v := range ids {
		userCmd["analytics:"+v] = tx.HGetAll(ctx, "analytics:"+v)
	}

	_, err = tx.Exec(ctx)
	if err != nil {
		return []UserMap{}, err
	}

	for _, v := range userCmd {
		var user UserMap
		err = v.Scan(&user)
		users = append(users, user)
	}

	return users, err
}

func (d *Dependency) FlushAllUserID(ctx context.Context) error {
	ids, err := d.GetAllUserID(ctx)
	if err != nil {
		return err
	}

	tx := d.Redis.TxPipeline()
	defer tx.Close()

	for _, v := range ids {
		tx.Del(ctx, "analytics:"+v)
	}

	tx.Del(ctx, "analytics:users")

	_, err = tx.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}
