# Publishing DiscordKit v0.1.0

The commits, the `main` merge and the annotated tag `v0.1.0` are all present in
this repository, but they could **not be pushed** from the build environment:
the available GitHub token is read-only for this repo (`git push` returns
`403 Permission ... denied to freitaseric`). Run the steps below from a machine
authenticated with push rights.

## Option A — push from your clone

```bash
git remote add release https://github.com/freitaseric/discordkit.git   # or use your existing origin
git fetch --all
# From this repo (which already has the merge + tag):
git push origin main
git push origin v0.1.0
```

## Option B — apply the bundle

`discordkit-v0.1.0.bundle` contains every branch and the `v0.1.0` tag.

```bash
git clone https://github.com/freitaseric/discordkit.git
cd discordkit
git pull /path/to/discordkit-v0.1.0.bundle main
git fetch /path/to/discordkit-v0.1.0.bundle 'refs/tags/*:refs/tags/*'
git push origin main
git push origin v0.1.0
```

## Then: GitHub Release + Go proxy

1. Create a GitHub Release for tag `v0.1.0` (Releases → Draft new release).
2. Warm the Go module proxy:

   ```bash
   GOPROXY=https://proxy.golang.org go list -m github.com/freitaseric/discordkit@v0.1.0
   ```

   The first request triggers indexing; retry after a minute if needed.
