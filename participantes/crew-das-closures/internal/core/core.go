package core

import (
	"encoding/json"

	"github.com/dyammarcano/crew-das-closures/internal/client/openrouter"
	"github.com/dyammarcano/crew-das-closures/internal/model"
)

type Core struct {
	client *openrouter.Client
	key    string
}

func NewCore(urlStr string, opts openrouter.Option) (*Core, error) {
	client := openrouter.NewClient(urlStr, opts)
	return &Core{client: client}, nil
}

func (c *Core) AskQuestion(question []byte) (*model.FindServiceResponse, error) {
	obj := &model.FindServiceRequest{}
	if err := json.Unmarshal(question, obj); err != nil {
		return nil, err
	}

	// TODO mock response
	return &model.FindServiceResponse{
		Data: &model.ServiceData{
			ServiceID:   11,
			ServiceName: "Telefones de seguradoras",
		},
	}, nil
}

func (c *Core) SetKey(key string) error {
	c.key = key
	return nil
}
