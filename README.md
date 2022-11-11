# internal
modules used for dittotrade services

Packages from this module are used among different dittotrade services.
Due problems with using it as a private repository for now it has been made public.
WARNING: don't put any secrets, passwords to this module!
It is used only as a way to share the CODE specific to dittotrade projects.

Should be imported into service repository with 
go get github.com/dittotrade/internal

When adding new functionality consider increase tag after commit:
git tag
git tag v1.3.0
git push
git push origin v1.3.1

Total 0 (delta 0), reused 0 (delta 0), pack-reused 0
To github.com:dittotrade/internal.git
* [new tag]         v1.3.1 -> v1.3.1

See https://go.dev/blog/publishing-go-modules for details on versioning