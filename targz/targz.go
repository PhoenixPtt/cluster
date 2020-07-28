package targz

// package main

// tar包实现了文件的打包功能,可以将多个文件或者目录存储到单一的.tar压缩文件中
// tar本身不具有压缩功能,只能打包文件或目录
import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"

	"os"
	"path"
)

// func main() {
// 	// showFileHeader()

// 	basePath := "/home/cetc15/下载/dockerDemo/testFile/tarFile/"

// 	// 压缩文件夹或文件
// 	// TarGz(basePath+"alpine", basePath+"压缩/alpine.tar.gz", true)
// 	// TarGz(basePath+"alpine", basePath+"压缩/alpine.tar", false)

// 	//解压文件夹或文件
// 	tarFile_c := basePath + "压缩/alpine.tar"
// 	untarPath_c := basePath + "解压"
// 	UnTarGz(tarFile_c, untarPath_c, false)
// 	// tarFile_c := basePath + "压缩/alpine.tar.gz"
// 	// untarPath_c := basePath + "解压"
// 	// UnTarGz(tarFile_c, untarPath_c, true)
// }

//Header Demo
func ShowFileHeader() {
	fileName := "/home/cetc15/下载/dockerDemo/tarDemo/alpine/alpine_3.12.0_amd.tar.gz"
	sfileInfo, err := os.Stat(fileName)
	if err != nil {
		fmt.Println("get file status Err,", err)
		return
	}
	header, err := tar.FileInfoHeader(sfileInfo, "")
	if err != nil {
		fmt.Println("get Header Info Err,", err)
		return
	}
	fmt.Println(header.Name)
	fmt.Println(header.Mode)
	fmt.Println(header.Uid)
	fmt.Println(header.Gid)
	fmt.Println(header.Size)
	fmt.Println(header.ModTime)
	fmt.Println(header.Typeflag)
	fmt.Println(header.Linkname)
	fmt.Println(header.Uname)
	fmt.Println(header.Gname)
	fmt.Println(header.Devmajor)
	fmt.Println(header.Devminor)
	fmt.Println(header.AccessTime)
	fmt.Println(header.ChangeTime)
	fmt.Println(header.Xattrs)
}

func Tar(tt io.Reader, destFilePath string) error {
	byteStrean, err := ioutil.ReadAll(tt)
	if err != nil {
		panic(err)
		return err
	}
	err = ioutil.WriteFile(destFilePath, byteStrean, 0644)
	if err != nil {
		panic(err)
		return err
	}

	return err
}

//srcDirPath 源文件路径
//destFilePath 压缩后到文件
//bGz:true表示打包同时压缩;false表示只打包不压缩
func TarGz(srcDirPath string, destFilePath string, bGz bool) error {
	fw, err := os.Create(destFilePath)
	if err != nil {
		return err
	}
	defer fw.Close()

	var tw *tar.Writer
	if bGz {
		gw := gzip.NewWriter(fw)
		defer gw.Close()

		tw = tar.NewWriter(gw)
	} else {
		tw = tar.NewWriter(fw)
	}
	defer tw.Close()

	f, err := os.Open(srcDirPath)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if fi.IsDir() {
		err = compressDir(srcDirPath, path.Base(srcDirPath), tw)
		if err != nil {
			return err
		}
	} else {
		err := compressFile(srcDirPath, fi.Name(), tw, fi)
		if err != nil {
			return err
		}
	}
	return nil
}

func compressDir(srcDirPath string, recPath string, tw *tar.Writer) error {
	dir, err := os.Open(srcDirPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	fis, err := dir.Readdir(0)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		curPath := srcDirPath + "/" + fi.Name()

		if fi.IsDir() {
			err = compressDir(curPath, recPath+"/"+fi.Name(), tw)
			if err != nil {
				return err
			}
		}

		err = compressFile(curPath, recPath+"/"+fi.Name(), tw, fi)
		if err != nil {
			return err
		}
	}
	return nil
}

func compressFile(srcFile string, recPath string, tw *tar.Writer, fi os.FileInfo) error {
	if fi.IsDir() {
		hdr := new(tar.Header)
		hdr.Name = recPath + "/"
		hdr.Typeflag = tar.TypeDir
		hdr.Size = 0
		hdr.Mode = int64(fi.Mode())
		hdr.ModTime = fi.ModTime()

		err := tw.WriteHeader(hdr)
		if err != nil {
			return err
		}
	} else {
		fr, err := os.Open(srcFile)
		if err != nil {
			return err
		}
		defer fr.Close()

		hdr := new(tar.Header)
		hdr.Name = recPath
		hdr.Size = fi.Size()
		hdr.Mode = int64(fi.Mode())
		hdr.ModTime = fi.ModTime()

		err = tw.WriteHeader(hdr)
		if err != nil {
			return err
		}

		_, err = io.Copy(tw, fr)
		if err != nil {
			return err
		}
	}
	return nil
}

//bGz:true表示解包压缩文件；false表示解包普通文件
func UnTarGz(srcFilePath string, destDirPath string, bGz bool) error {
	os.Mkdir(destDirPath, os.ModePerm)

	fr, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer fr.Close()

	var tr *tar.Reader

	if bGz {
		gr, err := gzip.NewReader(fr)
		if err != nil {
			return err
		}
		defer gr.Close()

		tr = tar.NewReader(gr)
	} else {
		tr = tar.NewReader(fr)
	}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}

		if hdr.Typeflag != tar.TypeDir {
			os.MkdirAll(destDirPath+"/"+path.Dir(hdr.Name), os.ModePerm)

			fw, _ := os.OpenFile(destDirPath+"/"+hdr.Name, os.O_CREATE|os.O_WRONLY, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			_, err = io.Copy(fw, tr)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
