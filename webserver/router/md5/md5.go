package mymd5

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

//MD5生成哈希值
func GetMD5HashCode(data []byte) string {
	//方法一：
	//创建一个使用MD5校验的hash，Hash接口的对象
	hash := md5.New()
	//输入数据
	hash.Write(data)
	//计算机出哈希值,返回数据data的MD5校验和
	bytes := hash.Sum(nil)
	//将字符串编码为16进制格式,返回字符串
	hashcode := hex.EncodeToString(bytes)

	//方法二：
	//返回的是长度为16的数组
	//bytes := md5.Sum(data)
	//将数组转为切片，转换成16进制，返回字符串
	//hashcode := hex.EncodeToString(bytes[:])

	//返回哈希值
	return hashcode
}

// 获取文件的MD5值
func GetFileMD5HashCode(file *os.File) string {
	//创建一个使用MD5校验的hash，Hash接口的对象
	hash := md5.New()
	// 将文件对象拷贝到哈希接口对象中
	io.Copy(hash, file)
	//计算机出哈希值,返回数据data的MD5校验和
	bytes := hash.Sum(nil)
	//将字符串编码为16进制格式,返回字符串
	hashcode := hex.EncodeToString(bytes)

	return hashcode
}

