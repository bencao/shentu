## certikd add-genesis-account

Add a genesis account to genesis.json

### Synopsis

Add a genesis account to genesis.json. The provided account must specify
the account address or key name and a list of initial coins. If a key name is given,
the address will be looked up in the local Keybase. The list of initial tokens must
contain valid denominations. Accounts may optionally be supplied with vesting parameters.
the precedence rule is period > continuous > endtime.


```
certikd add-genesis-account [address_or_key_name] [coin][,[coin]] [flags]
```

### Options

```
      --continuous                set to continuous vesting.
  -h, --help                      help for add-genesis-account
      --home-client string        client's home directory (default "~/.certikcli")
      --keyring-backend string    Select keyring's backend (os|file|test) (default "os")
      --manual                    set to manual vesting
      --num-periods uint          number of months for monthly vesting (default 1)
      --period uint               set to periodic vesting with period in seconds
      --unlocker string           address that can unlock this account's locked coins
      --vesting-amount string     amount of coins for vesting accounts
      --vesting-end-time uint     schedule end time (unix epoch) for vesting accounts
      --vesting-start-time uint   schedule start time (unix epoch) for vesting accounts
```

### Options inherited from parent commands

```
      --home string        directory for config and data (default "~/.certikd")
      --log_level string   Log level (default "main:info,state:info,*:error")
      --trace              print out full stack trace on errors
```

### SEE ALSO

* [certikd](certikd.md)	 - CertiK App Daemon (server)


