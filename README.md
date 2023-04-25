## What is this?
ethdont splits an existing BLS12-381 key using Shamir's secret sharing with a chosen threshold.
Each key share is then imported into a Dirk distributed wallet.

## Build

```
make build
```

## Run

```
# Create a random seed for the deterministic key generator
KEYGEN_SEED=$(head -c 32 /dev/urandom | xxd -p -c 32)
DIRK_PEERS="1:dirk1:13141,2:dirk2:13142,3:dirk3:13143"
DIRK_PEER_IDS="$(echo $DIRK_PEERS | awk -F '[,:]' '{for(i=1;i<=NF;i+=3) printf "%s ", $i; print ""}')"
for id in $DIRK_PEER_IDS; do
    ./ethdont \
        -dirk-peers "${DIRK_PEERS}" \
        -keystore keystores/keystore.json \
        -keystore-passphrase "$(cat keystores/secret)" \
        -threshold 2 \
        -wallet DistributedWallet \
        -wallet-passphrase "something secret" \
        -keygen-seed "${KEYGEN_SEED}" \
        -account validator-1 \
        -wallet-base-dir wallets/dirk${id}/wallets/ \
        -dirk-id "${id}"
done
```

### Output

```
Successfully imported account. Public key share: b3c05703271b3f36a8d21d8d9cf17d60756c2624584fb53a1e6145986c576f3bf081ca7e53d493fa02394b65857642fb, Public key: 8776e1020d2461884b92242ffcb5766fc8601a7e2ba9b3d213416a40d3dde8fe1dc018a8f95c64c0e99588046c505426
Successfully imported account. Public key share: 8e34fdb6364481dd2ab2e92b4f1d6a539f0208fc1b8a095ae4ebfa70a57344535da2abaa944f6b69af00a3cf7b91b832, Public key: 8776e1020d2461884b92242ffcb5766fc8601a7e2ba9b3d213416a40d3dde8fe1dc018a8f95c64c0e99588046c505426
Successfully imported account. Public key share: ae07cca7b5bb10d7d72ac41959a6f4ff36ac2c27fcebb0486375ae5e601d3936ba19b8855053934ad91221da7216c3a3, Public key: 8776e1020d2461884b92242ffcb5766fc8601a7e2ba9b3d213416a40d3dde8fe1dc018a8f95c64c0e99588046c505426
```
