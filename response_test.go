package xxljob_test

import (
	"encoding/json"
	"testing"

	"github.com/hyperjiang/xxljob"
	"github.com/stretchr/testify/require"
)

func TestResponse(t *testing.T) {
	should := require.New(t)

	s := []byte(`{"code":200,"msg":null,"content":null}`)

	var res xxljob.Response
	err := json.Unmarshal(s, &res)
	should.NoError(err)
	should.Equal(200, res.Code)
	should.Empty(res.Msg)

	should.Equal(`{"code":200,"msg":""}`, res.String())

	should.Equal(`{"code":200,"msg":""}`, xxljob.NewSuccResponse().String())
	should.Equal(`{"code":500,"msg":"error"}`, xxljob.NewErrorResponse("error").String())
}
