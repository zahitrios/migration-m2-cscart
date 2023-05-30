package services

import (
	"errors"
	"os"
)

const (
	gamaParam = "gredir=gama"
)

func DefineEnv(stage string) error {
	if stage == "stg" {
		os.Setenv("magentoUrl", os.Getenv("STG_MAGENTO_URL"))
		os.Setenv("magentoBearer", os.Getenv("STG_MAGENTO_BEARER"))
		os.Setenv("gamaUrl", os.Getenv("STG_GAMA_URL"))
		os.Setenv("gamaUser", os.Getenv("STG_GAMA_USERNAME"))
		os.Setenv("gamaPassword", os.Getenv("STG_GAMA_PASSWORD"))
		os.Setenv("gamaParam", gamaParam)
		return nil
	} else if stage == "prod" {
		os.Setenv("magentoUrl", os.Getenv("PROD_MAGENTO_URL"))
		os.Setenv("magentoBearer", os.Getenv("PROD_MAGENTO_BEARER"))
		os.Setenv("gamaUrl", os.Getenv("PROD_GAMA_URL"))
		os.Setenv("gamaUser", os.Getenv("PROD_GAMA_USERNAME"))
		os.Setenv("gamaPassword", os.Getenv("PROD_GAMA_PASSWORD"))
		os.Setenv("gamaParam", gamaParam)
		return nil
	} else {
		return errors.New("Stage " + stage + " is not defined")
	}
}
