package analytics

import (
	"context"
	"strings"

	"github.com/go-redis/redis/v8"
)

func (d *Dependency) GetAllUserID(ctx context.Context) ([]string, error) {
	r, err := d.Redis.Get(ctx, "analytics:users").Result()
	if err != nil {
		return []string{}, err
	}

	return strings.Split(r, " "), nil
}

func (d *Dependency) FlushAllUserID(ctx context.Context) error {
	ids, err := d.GetAllUserID(ctx)
	if err != nil {
		return err
	}

	tx := d.Redis.TxPipeline()
	defer tx.Close()

	for _, v := range ids {
		err = tx.Del(ctx, "analytics"+v).Err()
		if err != nil {
			return err
		}
	}

	err = tx.Set(ctx, "analytics:user", "", redis.KeepTTL).Err()
	if err != nil {
		return err
	}

	err = tx.Do(ctx).Err()
	if err != nil {
		return err
	}
	return nil
}
