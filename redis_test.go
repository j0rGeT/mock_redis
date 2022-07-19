package redis_core

import (
	"bou.ke/monkey"
	"errors"
	"github.com/alicebob/miniredis/v2"
	. "github.com/golang/mock/gomock"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/suite"
	mock_mocks "gitlab.archeros.cn/Haihe/go-core/redis_core/mocks"
	"log"
	"reflect"
	"testing"
	"time"
)

const mockerrorstring = "this is a mock error"

type RedisTestSuite struct {
	suite.Suite
	redisPool *RedisPool
	abnormalRedisPool *RedisPool
	redisServer *miniredis.Miniredis
}

func (suite *RedisTestSuite)InitRedisPoolServer() error {
	mr, err := miniredis.Run()
	mr.RequireAuth("123456")
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	var redisPool = &RedisPool{}
	var abnormalRedisPool = &RedisPool{}
	err = redisPool.New(&RedisConf{
		EndPoint: mr.Addr(),
		Password: "123456",
		IdleTimeout: 10000,
		ReadTimeout: 1000,
		WriteTimeout: 1000,
		DialConnectTimeout: 1000,
		TestOnBorrow: func(conn RedisConn, t time.Time) error {
			return nil
		},
	})

	_ = abnormalRedisPool.New(&RedisConf{
		EndPoint: mr.Addr(),
		Password: "",
		IdleTimeout: 10000,
		ReadTimeout: 1000,
		WriteTimeout: 1000,
		DialConnectTimeout: 1000,
		TestOnBorrow: func(conn RedisConn, t time.Time) error {
			return nil
		},
	})

	suite.redisPool = redisPool
	suite.abnormalRedisPool = abnormalRedisPool
	suite.redisServer = mr

	return err
}

func (suite *RedisTestSuite) SetupTest() {
	err := suite.InitRedisPoolServer()
	if err != nil {
		log.Fatal("init redis pool server failed")
	}
}

func (suite *RedisTestSuite) TestRedisCheckAlive() {
	rc := suite.redisPool.Client.Get()
	err := suite.redisPool.Client.TestOnBorrow(rc, time.Now())
	suite.Equal(err, nil)
}

func (suite *RedisTestSuite) TestSetWhenNormal() {
	err := suite.redisPool.Set("111", "222")
	suite.Equal(err , nil)
}

func (suite *RedisTestSuite) TestSetWhenAbnormal() {
	err := suite.abnormalRedisPool.Set("111", "222")
	suite.NotEqual(err, nil)
}


func (suite *RedisTestSuite) TestGetWhenNormal() {
	err := suite.redisPool.Set("111", "222")
	if err != nil {
		suite.Error(errors.New("redis set info failed"))
	}
	info, err := suite.redisPool.Get("111")
	suite.Equal(info, "222")
}

func (suite *RedisTestSuite) TestGetWhenAbnormal() {
	_, err := suite.abnormalRedisPool.Get("111")
	suite.NotEqual(err, nil)
}

func (suite *RedisTestSuite) TestGetWhenValueIsNil() {
	_, err := suite.redisPool.Get("111")
	suite.NotEqual(err, nil)
}


func (suite *RedisTestSuite) TestSetWithTTLWhenNormal() {
	err := suite.redisPool.SetWithTTL("111", 10, "333")
	suite.Equal(err, nil)
}

func (suite *RedisTestSuite) TestSetWithTTLWhenAbnormal() {
	err := suite.abnormalRedisPool.SetWithTTL("111", 10, "333")
	suite.NotEqual(err, nil)
}

func (suite *RedisTestSuite) TestHGetWhenNormal() {
	err := suite.redisPool.HMSet("hello", map[string]string{"222": "world"})
	if err != nil {
		suite.Error(errors.New("redis hget info failed"))
	}
	info, err := suite.redisPool.HGet("hello", "222")
	suite.Equal(info, "world")
}

func (suite *RedisTestSuite) TestHGetWhenAbnormal() {
	err := suite.redisPool.HMSet("hello", nil)
	if err != nil {
		suite.Error(errors.New("redis hset info failed"))
	}
	_, err = suite.redisPool.HGet("hello", "222")
	suite.NotEqual(err, nil)
}

func (suite *RedisTestSuite) TestHMSetWhenAbormal() {
	err := suite.abnormalRedisPool.HMSet("hello", map[string]string{"222": "world"})
	suite.NotEqual(err, nil)
}

func (suite *RedisTestSuite) TestHGetWhenAbormal() {
	_, err := suite.abnormalRedisPool.HGet("hello", "222")
	suite.NotEqual(err, nil)
}

func (suite *RedisTestSuite) TestHMSetWithTTLWhenNormal() {
	err := suite.redisPool.HMSetWithTTL("hello", 100, map[string]string{"222": "world"})
	suite.Equal(err, nil)
}

// gomock ways
// mock interface方式来mock Do接口方法
// monkey+gomock方式来mock，暂时没想到其他办法
// mocks.go生成方式： mockgen -source=mocks.go -destination repo.go生成
func (suite *RedisTestSuite) TestHMSetWithTTLWhenExpireFailed() {
	c := &redis.Pool{}
	ctrl := NewController(suite.T())
	defer ctrl.Finish()
	monkey.PatchInstanceMethod(reflect.TypeOf(c), "Get", func (p *redis.Pool) redis.Conn {
		mockRepo := mock_mocks.NewMockConn(ctrl)
		mockRepo.EXPECT().Do("HMSET", redis.Args{}.Add("111").AddFlat(map[string]string{"222": "world"})...).Return(0, nil)
		mockRepo.EXPECT().Do("EXPIRE", "111", 222).Return(nil, errors.New("this is a mocks"))
		mockRepo.EXPECT().Close().Return(nil)
		return mockRepo
	})
	err := suite.redisPool.HMSetWithTTL("111", 222, map[string]string{"222": "world"})
	suite.Equal(err, nil)
	monkey.UnpatchAll()
}

func (suite *RedisTestSuite) TestHMSetWithTTLWhenAbnormal() {
	err := suite.abnormalRedisPool.HMSetWithTTL("hello", 100, map[string]string{"222": "world"})
	suite.NotEqual(err, nil)
}

func (suite *RedisTestSuite) TestDelWhenNormal() {
	err := suite.redisPool.Set("111", "2")
	if err != nil {
		suite.Error(errors.New("redis hset info failed"))
	}
	err = suite.redisPool.Del("111")
	suite.Equal(err, nil)
}

func (suite *RedisTestSuite) TestDelWhenAbnormal() {
	err := suite.abnormalRedisPool.Del("111")
	suite.NotEqual(err, nil)
}

func (suite *RedisTestSuite) TestIsExistWhenNormal() {
	err := suite.redisPool.Set("111", "222")
	if err != nil {
		suite.Error(errors.New("redis is exists info normal failed"))

	}
	flag, _ := suite.redisPool.IsExist("111")
	suite.Equal(flag, true)
}

func (suite *RedisTestSuite) TestIsExistWhenAbnormal() {
	_, err := suite.abnormalRedisPool.IsExist("111")
	suite.NotEqual(err, nil)
}

func (suite *RedisTestSuite) TestIsExistWhenIntNil() {
	monkey.Patch(redis.Int, func (reply interface{}, err error) (int, error) {
		return 0, errors.New(mockerrorstring)
	})
	flag, err := suite.redisPool.IsExist("111")
	suite.Equal(flag, false)
	suite.NotEqual(err, nil)
	monkey.UnpatchAll()
}

func (suite *RedisTestSuite) TestIsExistWhenNotExist() {
	flag, _ := suite.redisPool.IsExist("aaa")
	suite.Equal(flag, false)
}


func (suite *RedisTestSuite) TestIndistinctDelWhenNormal() {
	err := suite.redisPool.Set("111", "222")
	if err != nil {
		suite.Error(errors.New("redis indistinct delete normal info failed"))
	}
	err = suite.redisPool.IndistinctDel("111")
	suite.Equal(err, nil)
}

func (suite *RedisTestSuite) TestIndistinctDelWhenAbnormal() {
	err := suite.abnormalRedisPool.IndistinctDel("111")
	suite.NotEqual(err, nil)
}
func (suite *RedisTestSuite) TestNewWhenDialIsError() {
	var c *RedisPool
	monkey.PatchInstanceMethod(reflect.TypeOf(c), "Dial", func (c *RedisPool, conf *RedisConf) error  {
		return errors.New(mockerrorstring)
	})
	defer monkey.UnpatchInstanceMethod(reflect.TypeOf(c), "Dial")
	err := suite.redisPool.New(&RedisConf{})
	suite.NotEqual(err, nil)
}

func (suite *RedisTestSuite) TestNewWhenRedisDialIsError() {
	monkey.Patch(redis.Dial, func (network, address string, options ...redis.DialOption) (redis.Conn, error) {
		return nil, errors.New(mockerrorstring)
	})
	defer monkey.Unpatch(redis.Dial)
	var redisPool = &RedisPool{}
	err := redisPool.New(&RedisConf{
		EndPoint: suite.redisServer.Addr(),
		Password: "123456",
		IdleTimeout: 10000,
		ReadTimeout: 1000,
		WriteTimeout: 1000,
		DialConnectTimeout: 1000,
		TestOnBorrow: nil,
	})
	suite.Equal(err, nil)
}

func (suite *RedisTestSuite) TestNewWhenRedisTestOnBorrowIsNil() {
	var redisPool = &RedisPool{}
	err := redisPool.New(&RedisConf{
		EndPoint: suite.redisServer.Addr(),
		Password: "123456",
		IdleTimeout: 10000,
		ReadTimeout: 1000,
		WriteTimeout: 1000,
		DialConnectTimeout: 1000,
		TestOnBorrow: nil,
	})
	rc := redisPool.Client.Get()
	err = redisPool.Client.TestOnBorrow(rc, time.Now())
	suite.Equal(err, nil)
}

func (suite *RedisTestSuite) TestNewWhenRedisAuthFailure() {
	var redisPool = &RedisPool{}
	err := redisPool.New(&RedisConf{
		EndPoint: suite.redisServer.Addr(),
		Password: "1234561111",
		IdleTimeout: 10000,
		ReadTimeout: 1000,
		WriteTimeout: 1000,
		DialConnectTimeout: 1000,
		TestOnBorrow: nil,
	})
	err = redisPool.Set("111", "222")
	suite.NotEqual(err, nil)
}

func (suite *RedisTestSuite) TestIndistinctDelWhenValuesIsNil() {
	monkey.Patch(redis.Values, func(reply interface{}, err error) ([]interface{}, error) {
		if reflect.TypeOf(reply).String() == "[]interface {}" {
			inf := reply.([]interface{})
			if len(inf) == 1 {
				return nil, errors.New(mockerrorstring)
			}else {
				return inf, nil
			}
		}
		return nil, errors.New(mockerrorstring)
	})
	defer monkey.UnpatchAll()
	suite.redisPool.Set("ggg", "ss")
	err := suite.redisPool.IndistinctDel("ggg")
	suite.NotEqual(err, nil)
}

func (suite *RedisTestSuite) TestMultiDel() {
	err := suite.redisPool.Set("111", "222")
	if err != nil {
		suite.Error(errors.New("redis multidel failed"))
	}
	err = suite.redisPool.MultiDel([]string{"111"})
	suite.Equal(err, nil)
}

func (suite *RedisTestSuite) TestMultiDelWhenAbnormal() {
	err := suite.abnormalRedisPool.MultiDel([]string{"111"})
	suite.NotEqual(err, nil)
}

func (suite *RedisTestSuite) TestSearch() {
	err := suite.redisPool.Set("111", "222")
	if err != nil {
		suite.Error(errors.New("redis search failed"))
	}
	res, err := suite.redisPool.Search("111")
	suite.Equal(err, nil)
	suite.Equal(res[0], "111")
}

func (suite *RedisTestSuite) TestSearchWhenAbnormal() {
	_, err := suite.abnormalRedisPool.Search("111")
	suite.NotEqual(err, nil)
}

func TestRedisSuite(t *testing.T) {
	suite.Run(t, new(RedisTestSuite))
}

func (suite *RedisTestSuite) TearDownTest() {
	suite.redisPool.Close()
}