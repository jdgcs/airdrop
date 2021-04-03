Aeternity AEX-9 token airdropping tool can be used for massively token airdropping independently in Windows, Linux and other OS.

# Usage
## Import wallet from Mnemonic
airdrop.exe import "glad bleak ... until assume axis verify" keystore_password keystore_name

## Airdrop
airdrop.exe keystore_name keystore_password contractid token_decimal airdroplist aenode 

example:
```shell
 airdrop.exe mywallet mypassword ct_BwJcRRa7jTAvkpzc2D16tJzHMGCJurtJMUBtyyfGi2QjPuMVv 16 lists.txt http://52.220.198.72:3013
```

and a .result file will be generated, such as **lists.txt.result**

## Check Airdrop
Check the result, for example,
```shell
airdrop.exe check lists.txt.result  http://52.220.198.72:3013
```

The failed transactions will be recorded in .err file, such as lists.txt.result.err, and the .err file can be used for re-airdropping.