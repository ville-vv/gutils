// Package valid
package valid

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type CustomClaims struct {
	Stash map[string]string
	jwt.StandardClaims
}

func (s *CustomClaims) SetStash(stash map[string]string) {
	if s.Stash == nil {
		s.Stash = make(map[string]string)
	}
	for key, val := range stash {
		s.Stash[key] = val
	}
}

type JwtWithRsa struct {
	PubKey []byte
	PrvKey []byte
}

// 使用Rsa签名token
func NewJwtWithRsa(pubKey, prvKey []byte) *JwtWithRsa {
	t := &JwtWithRsa{PubKey: pubKey, PrvKey: prvKey}
	copy(t.PrvKey, prvKey)
	copy(t.PubKey, pubKey)
	return t
}

// stash 可以存储用户自己的数据
// exp 过期时间 分钟
func (sel *JwtWithRsa) Generate(stash map[string]string, exp int64) (string, error) {
	claim := &CustomClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() + (exp * 60),
		},
	}
	// 存储自己的数据
	claim.SetStash(stash)
	token := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), claim)
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(sel.PrvKey)
	if err != nil {
		return "", err
	}
	return token.SignedString(privateKey)
}

func (sel *JwtWithRsa) Verify(dt string) (stash map[string]string, err error) {
	token, err := jwt.ParseWithClaims(dt, &CustomClaims{}, func(tk *jwt.Token) (i interface{}, e error) {
		publicKey, err := jwt.ParseRSAPublicKeyFromPEM(sel.PubKey)
		if err != nil {
			return nil, err
		}
		return publicKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			switch {
			case ve.Errors&jwt.ValidationErrorMalformed != 0:
				// ValidationErrorMalformed是一个uint常量，表示token不可用
				return nil, fmt.Errorf("token can not use")
			case ve.Errors&jwt.ValidationErrorExpired != 0:
				// ValidationErrorExpired表示Token过期
				return nil, fmt.Errorf("token expire")
			case ve.Errors&jwt.ValidationErrorNotValidYet != 0:
				// ValidationErrorNotValidYet表示无效token
				return nil, fmt.Errorf("invalid token")
			default:
				return nil, fmt.Errorf("token can not use")
			}
		}
		return nil, err
	}
	claim, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claim.Stash, nil
}
