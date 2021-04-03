//BSD license.
//Aeternity AEX-9 token airdropping tool, Ver 1.0
//Author: Liu Yang from www.aeknow.org,outcrop@163.com
//Usage: keystore_name keystore_password contractid token_decimal airdroplist aenode
//Airdrop: airdrop.exe mywallet mypassword ct_BwJcRRa7jTAvkpzc2D16tJzHMGCJurtJMUBtyyfGi2QjPuMVv 16 lists.txt http://52.220.198.72:3013
//Import account from Mnemonic: airdrop.exe import "glad bleak broccoli trip... until assume axis verify" password filename
//Check airdropping: airdrop.exe check lists.txt.result  http://52.220.198.72:3013

package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/aeternity/aepp-sdk-go/v8/account"
	//"github.com/aeternity/aepp-sdk-go/v8/cmd"
	aeconfig "github.com/aeternity/aepp-sdk-go/v8/config"
	"github.com/aeternity/aepp-sdk-go/v8/naet"
	"github.com/aeternity/aepp-sdk-go/v8/transactions"
	//"github.com/aeternity/aepp-sdk-go/v8/utils"
)

var ostype = runtime.GOOS

func main() {
	if os.Args[1] == "check" {
		checkfile := os.Args[2]
		aenode := os.Args[3]
		AirDropCheck_AEX9(checkfile, aenode)
	} else if os.Args[1] == "import" {
		mnemonic := os.Args[2]
		password := os.Args[3]
		filename := os.Args[4]
		ImportAccountFromMnemonic(mnemonic, password, filename)
	} else {
		if len(os.Args) < 7 {
			fmt.Println("Parameter error")
			os.Exit(1)
		}

		keystore := os.Args[1]
		password := os.Args[2]
		contractid := os.Args[3]
		token_decimal := os.Args[4]
		airdroplist := os.Args[5]
		aenode := os.Args[6]

		AirDrop_AEX9(keystore, password, contractid, token_decimal, airdroplist, aenode)

	}
}

//根据助记词导入
func ImportAccountFromMnemonic(mnemonic, password, filename string) {
	account_index := 0
	address_index := 0
	seed, err := account.ParseMnemonic(mnemonic)
	if err != nil {
		fmt.Println(err)
	}

	// Derive the subaccount m/44'/457'/3'/0'/1'
	key, err := account.DerivePathFromSeed(seed, uint32(account_index), uint32(address_index))
	if err != nil {
		fmt.Println(err)
	}

	// Deriving the aeternity Account from a BIP32 Key is a destructive process
	mykey, err := account.BIP32KeyToAeKey(key)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(mykey.Address)
	accountFileName := filename
	if !FileExist(accountFileName) {
		account.StoreToKeyStoreFile(mykey, password, accountFileName)
	} else {
		fmt.Println("File name exists.")
	}
}

//检测空投的tx是否成功，没成功的话记录到.err文件，准备下一次空投
func AirDropCheck_AEX9(checkfile, aenode string) {

	fi, err := os.Open(checkfile)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer fi.Close()

	content := ""
	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}

		airdropto := strings.Split(string(a), ":")
		txHash := airdropto[2]
		node := naet.NewNode(aenode, false)
		TopHeight, _ := node.GetHeight()

		v, err := node.GetTransactionByHash(txHash)

		if err != nil {
			fmt.Println("transaction error", err)
			content = content + string(a) + ":FAILED\n"
		} else {
			txheight, _ := strconv.Atoi(v.BlockHeight.String())

			if (TopHeight - uint64(txheight)) > 100 {
				fmt.Println("transaction OK", TopHeight, v, err)
			} else {
				if (TopHeight - uint64(txheight)) > 0 {
					fmt.Println("transaction Onchain", TopHeight, v, err)
				}
			}

			if txheight == -1 {
				fmt.Println("transaction in pool", txHash)
				content = content + string(a) + ":INPOOL\n"
			}

			if (TopHeight - uint64(txheight)) == 0 {
				fmt.Println("transaction to be mined", txHash)
			}
		}

	}

	if content != "" {
		resultfile := checkfile + ".err"
		f, err := os.Create(resultfile)
		defer f.Close()
		if err != nil {
			fmt.Println(err.Error())
		} else {
			_, err = f.Write([]byte(content))
		}
	}

}

//空投代币主程序
func AirDrop_AEX9(keystore, password, contractid, token_decimal, airdroplist, aenode string) {

	myAccount, err := account.LoadFromKeyStoreFile(keystore, password)
	fmt.Println(myAccount.Address)
	if err != nil {
		fmt.Println("Could not open the account:", err)
		os.Exit(1)
	}

	fi, err := os.Open(airdroplist)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer fi.Close()

	resultfile := "result.txt"
	content := ""
	posted := ""
	resultfile = airdroplist + ".result"
	// 存在文件则读取整个文件后，判断是否在其中；第一次空投，不存在则创建
	if FileExist(resultfile) {
		b, err := ioutil.ReadFile(resultfile)
		if err != nil {
			fmt.Print(err)
		}
		posted = string(b)
	} else {
		f, err := os.Create(resultfile)
		defer f.Close()
		if err != nil {
			fmt.Println(err.Error())
		} else {
			_, err = f.Write([]byte(""))
		}
	}

	//开始添加
	fappend, err := os.OpenFile(resultfile, os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err.Error())

	}
	defer fappend.Close()

	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		if !strings.Contains(posted, string(a)) {
			airdropto := strings.Split(string(a), ":")
			famount, err := strconv.ParseFloat(airdropto[1], 64)
			token_dec, err := strconv.Atoi(token_decimal)
			bigfloatAmount := big.NewFloat(famount)
			imultiple := big.NewFloat(math.Pow10(token_dec)) //N decimal
			//imultiple := big.NewFloat(1000000000000000000) //18 dec
			fmyamount := big.NewFloat(1)
			fmyamount.Mul(bigfloatAmount, imultiple)
			myamount := new(big.Int)
			fmyamount.Int(myamount)

			transferamount := myamount.String()

			node := naet.NewNode(aenode, false)
			ttlnoncer := transactions.NewTTLNoncer(node)
			ownerID := myAccount.Address

			abiVersion := uint16(3)
			amount := big.NewInt(0)
			gasLimit := big.NewInt(10000)
			gasPrice := big.NewInt(1000000000)
			callData := Contract_getCallData("transfer("+airdropto[0]+","+transferamount+")", "aex9.aes")
			fmt.Println("transfer(" + airdropto[0] + "," + transferamount + ")")

			TokenSignAccount := myAccount
			contractID := contractid
			txinfo := ""

			tx, err := transactions.NewContractCallTx(ownerID, contractID, amount, gasLimit, gasPrice, abiVersion, callData, ttlnoncer)
			if err != nil {
				fmt.Println("Could not create theTx:", err)
			} else {
				//fmt.Println(tx)
			}

			_, myTxhash, _, err := SignBroadcastTransaction(tx, TokenSignAccount, node, aeconfig.Node.NetworkID)
			if err != nil {
				fmt.Println("SignBroadcastTransaction failed with:", err)

			} else {
				txinfo = airdropto[0] + ":" + airdropto[1] + ":" + myTxhash
				content = txinfo + "\n"
				fmt.Println(txinfo)
				// 查找文件末尾的偏移量
				n, _ := fappend.Seek(0, 2)

				// 从末尾的偏移量开始写入内容
				_, err = fappend.WriteAt([]byte(content), n)

			}
			time.Sleep(time.Duration(3) * time.Second)
		}
	}

}

//编译获得代币调用的string
func Contract_getCallData(callStr, callContract string) string {

	if ostype == "windows" {
		c := "bin\\sophia\\erts\\bin\\escript.exe bin\\sophia\\aesophia_cli --create_calldata contracts\\deploy\\" + callContract + " --call " + callStr
		cmd := exec.Command("cmd", "/c", c)
		fmt.Println(c)
		out, _ := cmd.Output()
		callData := strings.Trim(strings.Replace(string(out), "Calldata:", "", 1), "\n")
		fmt.Println("Exec result:" + string(out))
		fmt.Println(callData)
		return callData
	} else {
		c := "./bin/sophia/erts/bin/escript ./bin/sophia/aesophia_cli --create_calldata ./contracts/deploy/" + callContract + " --call \"" + callStr + "\""
		fmt.Println(c)
		cmd := exec.Command("sh", "-c", c)
		out, _ := cmd.Output()
		callData := strings.Trim(strings.Replace(string(out), "Calldata:", "", 1), "\n")
		fmt.Println(callData)
		return callData
	}

	return ""

}

//广播tx到节点
func SignBroadcastTransaction(tx transactions.Transaction, signingAccount *account.Account, n naet.PostTransactioner, networkID string) (signedTxStr, hash, signature string, err error) {

	signedTx, hash, signature, err := transactions.SignHashTx(signingAccount, tx, networkID)
	if err != nil {
		return
	}
	fmt.Println(hash)
	signedTxStr, err = transactions.SerializeTx(signedTx)

	//fmt.Println(signedTxStr)
	if err != nil {
		return
	}

	err = n.PostTransaction(signedTxStr, hash)
	if err != nil {
		return
	}
	return
}

func FileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}
