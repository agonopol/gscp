package gscplib

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
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
	code    string
	service *storage.Service
	client  *http.Client
}

func NewStore(clientId, clientSecret, cache, code string) *store {
	this := new(store)
	this.code = code
	this.auth = &oauth.Config{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		Scope:        scope,
		AuthURL:      authURL,
		TokenURL:     tokenURL,
		TokenCache:   oauth.CacheFile(cache),
		RedirectURL:  redirectURL,
	}

	return this
}

func (this *store) getService() *storage.Service {
	this.init()
	return this.service
}

func (this *store) getClient() *http.Client {
	this.init()
	return this.client
}

func (this *store) init() {
	if this.service == nil || this.client == nil {
		// Set up a transport using the config
		transport := &oauth.Transport{
			Config:    this.auth,
			Transport: http.DefaultTransport,
		}

		token, err := this.auth.TokenCache.Token()
		if err != nil {
			if this.code == "" {
				url := this.auth.AuthCodeURL("")
				fmt.Println("Visit URL to get a code then run again with GSCP_CODE=YOUR_CODE")
				fmt.Println(url)
				os.Exit(1)
			}
			// Exchange auth code for access token
			token, err = transport.Exchange(this.code)
			if err != nil {
				log.Fatal("Exchange: ", err)
			}
			fmt.Printf("Token is cached in %v\n", this.auth.TokenCache)
		}
		transport.Token = token

		httpClient := transport.Client()
		this.client = httpClient
		this.service, err = storage.New(httpClient)
	}
}

func (this *store) Put(local, remote string) error {
	project, bucket, objectName := GetRemoteObjectName(SplitRemote(remote))
	if _, err := this.getService().Buckets.Get(bucket).Do(); err != nil {
		if _, err := this.getService().Buckets.Insert(project, &storage.Bucket{Name: bucket}).Do(); err != nil {
			return err
		}
	}
	object := &storage.Object{Name: objectName}
	file, err := os.Open(local)
	if err != nil {
		return err
	}
	_, err = this.getService().Objects.Insert(bucket, object).Media(file).Do()
	return err
}

func (this *store) download(remote, local string) error {

	err := os.MkdirAll(path.Dir(local), 0755)

	if !os.IsExist(err) && err != nil {
		log.Fatal(err)
	}
	out, err := os.Create(local)
	defer out.Close()

	resp, err := this.client.Get(remote)
	defer resp.Body.Close()

	n := int64(0)
	for n < resp.ContentLength {
		r, err := io.Copy(out, resp.Body)
		if err != nil {
			return err
		}
		n += r
	}
	return nil
}

func (this *store) Get(remote, local string) error {
	_, bucket, object := GetRemoteObjectName(SplitRemote(remote))
	if local == "." {
		local = object
	}
	if res, err := this.getService().Objects.Get(bucket, object).Do(); err == nil {
		return this.download(res.MediaLink, local)
	} else {
		return err
	}
}

func (this *store) Buckets(project string) ([]string, error) {
	this.getService().Buckets.List(project).Do()
	if res, err := this.getService().Buckets.List(project).Do(); err == nil {
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
	if res, err := this.getService().Objects.List(bucket).Do(); err == nil {
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
	_, err := this.getService().ObjectAccessControls.Insert(bucket, object, acl).Do()
	return err
}
