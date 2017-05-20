package mi_session

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/stormgbs/gopkg/utils"
)

var ErrorSignatureNotMatched = errors.New("signature not matched")

type UserSession struct {
	CasUsername string `redis:"CAS_USERNAME" json:"CAS_USERNAME"`
}

type RedisConfig struct {
	KeyPrefix string
	Host      string
	Port      int
	Password  string
}

type SessionStore struct {
	secretKey   string
	redisConfig *RedisConfig
	redisPool   *redis.Pool
}

func NewSessionStore(secretKey string, redisConfig *RedisConfig) *SessionStore {
	redisPool := utils.NewRedisConnPool(fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		redisConfig.Password)
	return &SessionStore{
		secretKey:   secretKey,
		redisConfig: redisConfig,
		redisPool:   redisPool,
	}
}

func getSessionId(session_string string) (string, error) {
	index := strings.LastIndex(session_string, signatureSeparator)
	if index < 0 {
		return "", fmt.Errorf("invalid session format: %s", session_string)
	}

	return session_string[:index], nil
}

func (s *SessionStore) verify(session_string string) (bool, error) {
	session_id, err := getSessionId(session_string)
	if err != nil {
		return false, err
	}

	sess := Session{s.secretKey}

	key, err := sess.Unsign(session_string)
	if err != nil {
		return false, err
	}

	if key == session_id {
		return true, nil
	}

	return false, nil
}

func (s *SessionStore) GetUser(session_string string) (*UserSession, error) {
	ok, err := s.verify(session_string)
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrorSignatureNotMatched
	}

	session_id, err := getSessionId(session_string)
	if err != nil {
		return nil, err
	}

	conn := s.redisPool.Get()
	defer conn.Close()

	str, err := redis.String(conn.Do("GET", s.redisConfig.KeyPrefix+session_id))
	if err != nil {
		return nil, err
	}

	var v UserSession
	if err = json.Unmarshal([]byte(str), &v); err != nil {
		return nil, err
	}

	return &v, nil
}
