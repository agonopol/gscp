gscp
====

SCP Like commandline/lib for Google Cloud Storage


usage: 


GSCP_CLIENT_ID=CLIENT_ID GSCP_CLIENT_SECRET=CLIENT_SECRET gscp project@bucket:object file

gscp file project@bucket:object

gscp project@bucket

gscp project


gscplib:

store := NewStore(clientId, clientSecret, cache, code string)

store.Get(remote, local string) error

store.Put(local, remote string) error

store.Ls(bucket) (objects []string, err error)

store.Buckets(project) (buckets []string, err error)