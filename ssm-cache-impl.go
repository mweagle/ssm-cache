package ssmcache

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

type ssmCacheImpl struct {
	ssmSvc *ssm.SSM
	cache  *gocache.Cache
}

var _ Client = &ssmCacheImpl{}

func (ssmCache *ssmCacheImpl) getParameter(paramName string,
	encrypted bool,
	paramType string,
	expiry time.Duration) (string, error) {

	paramInput := &ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(encrypted),
	}

	paramValueOutput, paramValueOutputErr := ssmCache.ssmSvc.GetParameter(paramInput)
	if paramValueOutputErr != nil {
		return "", paramValueOutputErr
	}
	if *paramValueOutput.Parameter.Type != paramType {
		return "", errors.Errorf("Parameter %s is type %s, not type %s",
			paramName,
			*paramValueOutput.Parameter.Type,
			paramType)
	}
	paramValue := ""
	if paramValueOutput.Parameter.Value != nil {
		paramValue = *paramValueOutput.Parameter.Value
	}
	ssmCache.cache.Set(paramName, paramValue, expiry)
	return paramValue, nil
}

func (ssmCache *ssmCacheImpl) Purge(paramName string) Client {
	ssmCache.cache.Delete(paramName)
	return ssmCache
}

func (ssmCache *ssmCacheImpl) GetString(paramName string) (string, error) {
	return ssmCache.GetExpiringString(paramName, gocache.DefaultExpiration)
}

func (ssmCache *ssmCacheImpl) GetExpiringString(paramName string, expiry time.Duration) (string, error) {
	value, valueExists := ssmCache.cache.Get(paramName)
	if valueExists {
		stringValue, stringValueOk := value.(string)
		if !stringValueOk {
			return "", errors.Errorf("Failed to type assert cached param: %s", paramName)
		}
		return stringValue, nil
	}
	newValue, newValueErr := ssmCache.getParameter(paramName,
		false,
		ssm.ParameterTypeString,
		expiry)
	if newValueErr != nil {
		return "", errors.Wrapf(newValueErr, "Attempting to get parameter: %s", paramName)
	}
	return newValue, nil
}
func (ssmCache *ssmCacheImpl) GetStringList(paramName string) ([]string, error) {
	return nil, nil
}

func (ssmCache *ssmCacheImpl) GetExpiringStringList(paramName string, expiry time.Duration) ([]string, error) {
	value, valueExists := ssmCache.cache.Get(paramName)
	if valueExists {
		stringValue, stringValueOk := value.(string)
		if !stringValueOk {
			return nil, errors.Errorf("Failed to type assert cached param: %s", paramName)
		}
		return strings.Split(stringValue, ","), nil
	}
	newValue, newValueErr := ssmCache.getParameter(paramName,
		false,
		ssm.ParameterTypeStringList,
		expiry)
	if newValueErr != nil {
		return nil, errors.Wrapf(newValueErr, "Attempting to get parameter: %s", paramName)
	}
	return strings.Split(newValue, ","), nil
}

func (ssmCache *ssmCacheImpl) GetSecureString(paramName string) (string, error) {
	return ssmCache.GetExpiringSecureString(paramName, gocache.DefaultExpiration)
}

func (ssmCache *ssmCacheImpl) GetExpiringSecureString(paramName string, expiry time.Duration) (string, error) {
	value, valueExists := ssmCache.cache.Get(paramName)
	if valueExists {
		stringValue, stringValueOk := value.(string)
		if !stringValueOk {
			return "", errors.Errorf("Failed to type assert cached param: %s", paramName)
		}
		return stringValue, nil
	}
	newValue, newValueErr := ssmCache.getParameter(paramName,
		true,
		ssm.ParameterTypeSecureString,
		expiry)
	if newValueErr != nil {
		return "", errors.Wrapf(newValueErr, "Attempting to get parameter: %s", paramName)
	}
	return newValue, nil
}

func (ssmCache *ssmCacheImpl) GetParameterGroup(groupKey string, ssmKeyPath string) (ParameterGroup, error) {
	return ssmCache.GetExpiringParameterGroup(groupKey, ssmKeyPath, gocache.DefaultExpiration)
}

func (ssmCache *ssmCacheImpl) GetExpiringParameterGroup(groupKey string,
	ssmKeyPath string,
	expiry time.Duration) (ParameterGroup, error) {

	value, valueExists := ssmCache.cache.Get(groupKey)
	if valueExists {
		paramGroup, paramGroupOK := value.(ParameterGroup)
		if !paramGroupOK {
			return nil, errors.Errorf("Failed to type assert cached ParameterGroup type: %s", groupKey)
		}
		return paramGroup, nil
	}

	// Get all the parameters, stuff them into a map
	paramByPathInput := &ssm.GetParametersByPathInput{
		Path:      aws.String(ssmKeyPath),
		Recursive: aws.Bool(true),
	}
	// Walk the pages...
	paramMap := make(map[string]interface{}, 0)
	pagingErr := ssmCache.ssmSvc.GetParametersByPathPages(paramByPathInput,
		func(page *ssm.GetParametersByPathOutput, lastPage bool) bool {
			for _, eachParam := range page.Parameters {
				value := ""
				if eachParam.Value != nil {
					value = *eachParam.Value
				}
				paramMap[*eachParam.Name] = value
			}
			return true
		})

	if pagingErr != nil {
		return nil, errors.Wrapf(pagingErr, "Failed to page through all parameters in SSM tree: %s", ssmKeyPath)
	}
	// Cache it...
	ssmCache.cache.Set(groupKey, paramMap, expiry)
	return paramMap, nil
}
