package crypto

import (
	"crypto"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestCrypto(t *testing.T) {
	s := []byte("longalongalongha")
	k := []byte("hello world long")

	r, err := AESEncrypt(s, k)
	fmt.Printf("err : %s", err)

	// fmt.Println("r : ", string(r))

	res, err := AESDecrypt(r, k)
	fmt.Printf("err : %s", err)

	fmt.Println("res : ", string(res))
}

func TestRsa(t *testing.T) {
	text := "上山打老虎"
	usePKCS8 := true // usePKCS8=true表示是否成PKCS8格式的公私秘钥,否则乘车PKCS1格式的公私秘钥
	path, _ := os.Executable()
	filePath := filepath.Dir(path)
	fmt.Printf("文件路径：%s\n", filePath) // 存放pem,crt,pfx等文件的目录

	//生成Rsa
	publicKey, privateKey := GenerateRsaKey(usePKCS8)
	//从Pem文件读取秘钥，filePath是文件目录
	//publicKey, _ := ReadFromPem(filepath.Join(filePath, "rsa.pub"))
	//privateKey, _ := ReadFromPem(filepath.Join(filePath, "rsa.pem"))
	//从pfx文件中读取秘钥，filePath是文件目录
	//publicKey, privateKey := ReadFromPfx(filepath.Join(filePath, "demo.pfx"), "123456", usePKCS8)
	//从crt文件中读取公钥，filePath是文件目录
	//publicKey, _ := ReadPublicKeyFromCrt(filepath.Join(filePath, "demo.crt"), usePKCS8)
	//privateKey, _ := ReadFromPem(filepath.Join(filePath, "demo.key"))

	//保存到Pem文件，filePath是文件目录
	WriteToPem(false, publicKey, filepath.Join(filePath, "rsa.pub"))
	WriteToPem(true, privateKey, filepath.Join(filePath, "rsa.pem"))

	//Pkcs8格式公钥转换为Pkcs1格式公钥
	publicKey = Pkcs8ToPkcs1(false, publicKey)
	// Pkcs8格式私钥转换为Pkcs1格式私钥
	privateKey = Pkcs8ToPkcs1(true, privateKey)
	// Pkcs1格式公钥转换为Pkcs8格式公钥
	publicKey = Pkcs1ToPkcs8(false, publicKey)
	// Pkcs1格式私钥转换为Pkcs8格式私钥
	privateKey = Pkcs1ToPkcs8(true, privateKey)

	encryptText, _ := RsaEncrypt(text, publicKey, usePKCS8)
	fmt.Printf("【%s】经过【RSA】加密后：%s\n", text, encryptText)

	decryptText, _ := RsaDecrypt(encryptText, privateKey, usePKCS8)
	fmt.Printf("【%s】经过【RSA】解密后：%s\n", encryptText, decryptText)

	signature, _ := Sign(text, privateKey, crypto.MD5, usePKCS8)
	fmt.Printf("【%s】经过【RSA】签名后：%s\n", text, signature)

	result := Verify(text, publicKey, signature, crypto.MD5, usePKCS8) == nil
	fmt.Printf("【%s】的签名【%s】经过【RSA】验证后结果是："+strconv.FormatBool(result), text, signature)

}
