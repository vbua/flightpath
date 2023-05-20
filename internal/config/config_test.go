package config

import (
	"os"
	"testing"

	jsonIter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

func TestParseConfig_ConfigPathNotSet(t *testing.T) {
	t.Setenv(configPath, "")
	got, err := ParseConfig()
	assert.Nil(t, got)
	assert.EqualError(t, err, "env not set: CONFIG_PATH")
}

func TestParseConfig_NotExistingConfigFile(t *testing.T) {
	t.Setenv(configPath, "bad")
	got, err := ParseConfig()
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "can't read config file: open", "bad: no such file or directory")
}

func TestParseConfig_BadJsonUnmarshalError(t *testing.T) {
	t.Setenv(configPath, "bad.json")
	err := os.WriteFile("bad.json", []byte("bad"), os.ModePerm)
	assert.NoError(t, err)

	got, err := ParseConfig()
	assert.Nil(t, got)
	assert.EqualError(t, err, "can't unmarshal config: readObjectStart: "+
		"expect { or n, but found b, error found in #1 byte of ...|bad|..., bigger context ...|bad|...")

	err = os.Remove("bad.json")
	assert.NoError(t, err)
}

func TestParseConfig_Ok(t *testing.T) {
	t.Setenv(configPath, "full.json")
	cfg := Config{
		ServerOpts: ServerOpts{
			ReadTimeout:          10,
			WriteTimeout:         15,
			IdleTimeout:          20,
			MaxRequestBodySizeMb: 25,
		},
		MainPort: "1234",
	}

	buf, err := jsonIter.Marshal(cfg)
	assert.NoError(t, err)

	err = os.WriteFile("full.json", buf, os.ModePerm)
	assert.NoError(t, err)

	got, err := ParseConfig()
	assert.NoError(t, err)

	assert.Equal(t, &Config{
		ServerOpts: ServerOpts{
			ReadTimeout:          10,
			WriteTimeout:         15,
			IdleTimeout:          20,
			MaxRequestBodySizeMb: 25,
		},
		MainPort: "1234",
	}, got)

	err = os.Remove("full.json")
	assert.NoError(t, err)
}
