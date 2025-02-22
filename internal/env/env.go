package env

import (
	"os"
	"strconv"
)

func GetString(key ,fallback string)string{
	value , found :=os.LookupEnv(key);

	if(!found){
		return fallback
	}

	return value;
}

func GetInt(key string , fallback int) int{
	val, found :=os.LookupEnv(key)

	if(!found){
		return fallback;
	}

	valAsInt,err := strconv.Atoi(val)

	if(err!=nil){
		return fallback
	}

	return  valAsInt;
}