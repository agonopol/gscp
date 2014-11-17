package gscplib

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/google-api-go-client/storage/v1"
)

const (
	scope       = storage.DevstorageFull_controlScope
	authURL     = "https://accounts.google.com/o/oauth2/auth"
	tokenURL    = "https://accounts.google.com/o/oauth2/token"
	entityName  = "allUsers"
	redirectURL = "urn:ietf:wg:oauth:2.0:oob"
)

func SplitRemote(remote string) []string {
	fields := strings.FieldsFunc(
		remote, func(r rune) bool {
			return r == '@' || r == ':'
		})
	return fields
}

func GetRemoteObjectName(remote []string) (string, string, string) {
	if len(remote) < 3 {
		log.Fatalf("Remote path [%v] not fully specified (project@bucket:object)", remote)
	}
	return remote[0], remote[1], remote[2]
}

func GetRemoteBucketName(remote []string) (string, string) {
	if len(remote) < 2 {
		log.Fatalf("Remote bucket [%v] not fully specified (project@bucket)", remote)
	}
	return remote[0], remote[1]
}

type store struct {
	auth    *oauth.Config
	service *storage.Service
}

func NewStore(clientId, clientSecret, cacheFile string) *store {
	this := new(store)
	this.auth = &oauth.Config{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		Scope:        scope,
		AuthURL:      authURL,
		TokenURL:     tokenURL,
		TokenCache:   oauth.CacheFile(cacheFile),
		RedirectURL:  redirectURL,
	}
	return this
}

func (this *store) getService(code string) *storage.Service {
	if this.service == nil {
		// Set up a transport using the config
		transport := &oauth.Transport{
			Config:    this.auth,
			Transport: http.DefaultTransport,
		}

		token, err := this.auth.TokenCache.Token()
		if err != nil {
			if code == "" {
				url := this.auth.AuthCodeURL("")
				fmt.Println("Visit URL to get a code then run again with -code=YOUR_CODE")
				fmt.Println(url)
				os.Exit(1)
			}
			// Exchange auth code for access token
			token, err = transport.Exchange(code)
			if err != nil {
				log.Fatal("Exchange: ", err)
			}
			fmt.Printf("Token is cached in %v\n", this.auth.TokenCache)
		}
		transport.Token = token

		httpClient := transport.Client()
		this.service, err = storage.New(httpClient)
	}
	return this.service
}

func (this *store) Put(local, remote string) error {
	project, bucket, objectName := GetRemoteObjectName(SplitRemote(remote))
	if _, err := this.getService("").Buckets.Get(bucket).Do(); err != nil {
		if _, err := this.getService("").Buckets.Insert(project, &storage.Bucket{Name: bucket}).Do(); err != nil {
			return err
		}
	}
	object := &storage.Object{Name: objectName}
	file, err := os.Open(local)
	if err != nil {
		return err
	}
	_, err = this.getService("").Objects.Insert(bucket, object).Media(file).Do()
	return err
}

func (this *store) Get(remote, local string) error {
	// Get an object from a bucket.
	_, bucket, object := GetRemoteObjectName(SplitRemote(remote))
	if res, err := this.getService("").Objects.Get(bucket, object).Do(); err == nil {
		fmt.Printf("The media download link for %v/%v is %v.\n\n", bucket, res.Name, res.MediaLink)
		return nil
	} else {
		return err
	}
}

func (this *store) Buckets(project string) ([]string, error) {
	if res, err := this.getService("").Buckets.List(project).Do(); err == nil {
		buckets := make([]string, 0)
		for _, item := range res.Items {
			buckets = append(buckets, item.Id)
		}
		return buckets, nil
	} else {
		return []string{}, err
	}
}

func (this *store) Ls(remote string) ([]string, error) {
	_, bucket := GetRemoteBucketName(SplitRemote(remote))
	if res, err := this.getService("").Objects.List(bucket).Do(); err == nil {
		objects := make([]string, 0)
		for _, object := range res.Items {
			objects = append(objects, object.Name)
		}
		return objects, nil
	} else {
		return []string{}, err
	}
}

func (this *store) Chmod(remote, entity, role string) error {
	_, bucket, object := GetRemoteObjectName(SplitRemote(remote))
	acl := &storage.ObjectAccessControl{
		Bucket: bucket, Entity: entity, Object: object, Role: role,
	}
	_, err := this.getService("").ObjectAccessControls.Insert(bucket, object, acl).Do()
	return err
}
