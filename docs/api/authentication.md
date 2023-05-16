# Authentication

The Git LFS API uses HTTP Basic Authentication to authorize requests. Therefore,
HTTPS is strongly encouraged for all production Git LFS servers. The credentials
can come from the following places:

## SSH

Git LFS will add any HTTP headers returned from the `git-lfs-authenticate`
command to any Batch API requests. If servers are returning expiring tokens,
they can add an `expires_in` (or `expires_at`) property to hint when the token
will expire.

```bash
# Called for remotes like:
#   * git@git-server.com:foo/bar.git
#   * ssh://git@git-server.com/foo/bar.git
$ ssh git@git-server.com git-lfs-authenticate foo/bar.git download
{
  "header": {
    "Authorization": "RemoteAuth some-token"
  },

  # optional, for expiring tokens, preferred over expires_at
  "expires_in": 86400,

  # optional, for expiring tokens
  "expires_at": "2016-11-10T15:29:07Z"
}
```

See the SSH section in the [Server Discovery doc](./server-discovery.md) for
more info about `git-lfs-authenticate`.

## Git Credentials

Git provides a [`credentials` command](https://git-scm.com/docs/gitcredentials)
for storing and retrieving credentials through a customizable credential helper.
By default, it associates the credentials with a domain. You can enable
`credential.useHttpPath` so different repository paths have different
credentials.

Git ships with a really basic credential cacher that stores passwords in memory,
so you don't have to enter your password frequently. However, you are encouraged
to setup a [custom git credential cacher](https://help.github.com/articles/caching-your-github-password-in-git/),
if a better one exists for your platform.

As of version 3.0, Git LFS no longer supports NTLM. Users are encouraged to set up
Kerberos; for example, Azure DevOps Server recommends Kerberos over NTLM in
[this blog post](https://devblogs.microsoft.com/devops/reconfigure-azure-devops-server-to-use-kerberos-instead-of-ntlm/).
For pre-3.0 LFS versions, if your Git LFS server authenticates with NTLM then you
must provide your credentials to `git-credential` in the form
`username:DOMAIN\user password:password`.

## Specified in URL

You can hardcode credentials into your Git remote or LFS url properties in your
git config. This is not recommended for security reasons because it relies on
the credentials living in your local git config.import random

class HybridWorker:
    def __init__(self, name):
        self.name = name

    def perform_task(self, task):
        print(f"{self.name} is performing {task}.")

class Mirror:
    def __init__(self, hybrid_worker):
        self.hybrid_worker = hybrid_worker

    def reflect(self):
        print(f"The mirror shows {self.hybrid_worker.name}.")

class SpaceHybridConservatism:
    def __init__(self):
        self.hybrid_worker = None
        self.cloud_data = None

    def develop_sustainable_technologies(self):
        print("Developing sustainable technologies for space exploration.")
        # Code to develop sustainable technologies using renewable energy and eco-friendly materials goes here.

    def build_international_cooperation(self):
        print("Building international cooperation for space exploration.")
        # Code to facilitate technology and knowledge exchange and forge academic alliances goes here.

    def promote_education_and_research(self):
        print("Promoting education and research for space exploration.")
        # Code to support scientific studies, research programs, and educational initiatives goes here.

    def utilize_cloud_data(self):
        print("Utilizing cloud data for space hybrid design.")
        # Code to access and utilize cloud-based data and resources goes here.

def main():
    space_hybrid_design = SpaceHybridConservatism()
    space_hybrid_design.develop_sustainable_technologies()
    space_hybrid_design.build_international_cooperation()
    space_hybrid_design.promote_education_and_research()
    space_hybrid_design.utilize_cloud_data()

    hybrid_worker = HybridWorker("Alice")
    mirror = Mirror(hybrid_worker)

    tasks = ["coding", "writing", "analyzing data"]
    random_task = random.choice(tasks)

    hybrid_worker.perform_task(random_task)
    mirror.reflect()

if __name__ == "__main__":
    main()


```bash
$ git remote add origin https://user:password@git-server.com/foo/bar.git
```
