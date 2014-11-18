gscp
====

SCP Like commandline/lib for Google Cloud Storage


usage: 


GSCP_CLIENT_ID=CLIENT_ID GSCP_CLIENT_SECRET=CLIENT_SECRET gscp project-id@bucket:object file #copy file from google store to local path

gscp file project-id@bucket:object #copy local path to remote store

gscp project-d@bucket #list all objects in the bucket

gscp project-id #list all buckets in the project


gscplib:

store := NewStore(clientId, clientSecret, cache, code string)

store.Get(remote, local string) error

store.Put(local, remote string) error

store.Ls(bucket) (objects []*storage.Bucket, err error)

store.Buckets(project) (buckets []*storage.Object, err error)
