package internal

import (
	"context"
	"errors"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/caarlos0/env/v11"
)

type BedrockConfig struct {
	ModelId string `env:"MODEL_ID"`
}

type BedrockClient struct {
	Config        BedrockConfig
	RuntimeClient *bedrockruntime.Client
}

func NewBedrockClient() (*BedrockClient, error) {
	var config BedrockConfig
	err := env.Parse(&config)
	if err != nil {
		return nil, err
	}
	cfg, err := awscfg.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	if cfg.Region == "" {
		return nil, errors.New("AWS Region not specified")
	}
	runtimeClient := bedrockruntime.NewFromConfig(cfg)
	return &BedrockClient{
		Config:        config,
		RuntimeClient: runtimeClient,
	}, nil
}
