package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder"
	rtmtokenbuilder2 "github.com/AgoraIO-Community/go-tokenbuilder/rtmtokenbuilder"
	"github.com/gin-gonic/gin"
)

var appID, appCertificate string

func main() {
	appIDEnv, appIDExists := os.LookupEnv("APP_ID")
	appCertificateEnv, appCertificateExists := os.LookupEnv("APP_CERTIFICATE")

	if !appIDExists || !appCertificateExists {
		log.Fatal("Error: APP_ID and APP_CERTIFICATE environment variables are required.")
	} else {
		appID = appIDEnv
		appCertificate = appCertificateEnv
	}
	api := gin.Default()
	api.GET("/ping", func (c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	api.GET("rtc/:channelName/:role/:tokenType/:uid", getRtcToken)
	api.GET("rtm/:uid/", getRtmToken)
	api.GET("rte/:channelName/:role/:tokenType/:uid/", getBothRokens)

	api.Run(":8080")
}

func getRtcToken(c *gin.Context) {
	// get param values
	channelName, tokenType, uidStr, role, expireTimestamp, err := parseRtcParams(c)
	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(400,
		gin.H{"message": "Error Generating RTC token: " + err.Error(),
		"status": 400,
	})
		return
	}
	// generate the token
	rtcToken, tokenErr := generateRtcToken(channelName, uidStr, tokenType, role, expireTimestamp)
	if tokenErr != nil {
		log.Println("Error generating RTC token: ", tokenErr)
		c.Error(tokenErr)
		c.AbortWithStatusJSON(400, gin.H{
			"status": 400,
			"error": "Error generating RTC token: " + tokenErr.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"rtcToken": rtcToken,
		})
	}
	// return the token in JSON response
}

func getRtmToken(c *gin.Context) {
	// get param values
	uidStr, expireTimestamp, err := parseRtmParams(c)
	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(400, gin.H{
			"status": 400,
			"message": "Error Generating RTM token: " + err.Error(),
		})
		return
	}
	// build rtm token
	rtmToken, tokenErr := rtmtokenbuilder2.BuildToken(appID, appCertificate, uidStr, expireTimestamp, "")
	// return rtm token
	if tokenErr != nil {
		log.Println(err)
		c.Error(tokenErr)
		errMsg := "Error generating RTM token: " + tokenErr.Error()
		c.AbortWithStatusJSON(400, gin.H{
			"status": 400,
			"error": errMsg,
		})
		return
	}
	c.JSON(200, gin.H{
		"rtmToken": rtmToken,
	})
}

func getBothRokens(c *gin.Context) {
}

func parseRtcParams(c *gin.Context) (channelName, tokenType, uidStr string, role rtctokenbuilder2.Role, expireTimestamp uint32, err error) {
	// get param values
	channelName = c.Param("channelName")
	roleStr := c.Param("role")
	tokenType = c.Param("tokenType")
	uidStr = c.Param("uid")
	expireTime := c.DefaultQuery("expiry","3600")
	
	if roleStr == "publisher" {
		role = rtctokenbuilder2.RolePublisher
	} else {
		role = rtctokenbuilder2.RoleSubscriber
	}

	expireTime64, parseErr := strconv.ParseUint(expireTime, 10, 64)
	if parseErr != nil {
		err = fmt.Errorf("failed to parse expireTime: %s, causing erro: %s", expireTime, parseErr)
	}
	expireTimeInSeconds := uint32(expireTime64)
	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp = currentTimestamp + expireTimeInSeconds

	return channelName, tokenType, uidStr, role, expireTimestamp, err
}

func parseRtmParams(c *gin.Context) (uidStr string, expireTimestamp uint32, err error) {
	// get param values
	uidStr = c.Param("uid")
	expireTime := c.DefaultQuery("expiry","3600")

	expireTime64, parseErr := strconv.ParseUint(expireTime, 10, 64)
	if parseErr != nil {
		err = fmt.Errorf("failed to parse expireTime: %s, causing erro: %s", expireTime, parseErr)
	}
	expireTimeInSeconds := uint32(expireTime64)
	currentTimestamp := uint32(time.Now().UTC().Unix())
	expireTimestamp = currentTimestamp + expireTimeInSeconds

	return uidStr, expireTimestamp, err
}

func generateRtcToken(channelName, uidStr, tokenType string, role rtctokenbuilder2.Role, expireTimestamp uint32) (rtcToken string, err error) {
	// check token type
	if tokenType == "userAccount" {
		rtcToken, err = rtctokenbuilder2.BuildTokenWithAccount(appID, appCertificate, channelName, uidStr, role, expireTimestamp)
		return rtcToken, err
	} else if tokenType == "uid" {
		uid64, parseErr := strconv.ParseUint(uidStr, 10, 64)
		if parseErr != nil {
			err = fmt.Errorf("Failed to parse uidStr: %s, to uint causing error: %s", uidStr, parseErr)
			return "", err
		}
		uid := uint32(uid64)
		rtcToken, err = rtctokenbuilder2.BuildTokenWithUid(appID, appCertificate, channelName, uid, role, expireTimestamp)
		return rtcToken, err
	} else {
		err = fmt.Errorf("failed to generate RTC token for unknown tokenType: %s", tokenType)
		log.Println(err)
		return "", err
	}
}