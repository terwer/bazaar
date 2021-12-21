// SiYuan community bazaar.
// Copyright (c) 2021-present, b3log.org
//
// Pipe is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//         http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/88250/gulu"
	"github.com/parnurzeal/gorequest"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

var logger = gulu.Log.NewLogger(os.Stdout)

func main() {
	logger.Infof("bazaar is indexing...")
	indexBazaar()
	logger.Infof("indexed bazaar")
}

func indexBazaar() {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	data, err := cmd.CombinedOutput()
	if nil != err {
		logger.Fatalf("get git hash failed: %s", err)
	}
	hash := string(data)
	repoURL := "siyuan-note/bazaar@" + hash
	logger.Infof("bazaar [%s]", repoURL)
	data = downloadRepoZip(repoURL)

	bucket := os.Getenv("QINIU_BUCKET")
	ak := os.Getenv("QINIU_AK")
	sk := os.Getenv("QINIU_SK")

	key := time.Now().Format("bazaar/" + repoURL)
	putPolicy := storage.PutPolicy{
		Scope: fmt.Sprintf("%s:%s", bucket, key), // overwrite if exists
	}
	cfg := storage.Config{}
	cfg.Zone = &storage.ZoneHuanan
	formUploader := storage.NewFormUploader(&cfg)
	if err = formUploader.Put(context.Background(), nil, putPolicy.UploadToken(qbox.NewMac(ak, sk)),
		key, bytes.NewReader(data), int64(len(data)), nil); nil != err {
		logger.Fatalf("upload index [%s] failed: %s", hash, err)
	}

	logger.Infof("uploaded package [%s]", repoURL)
}

func downloadRepoZip(repoURL string) []byte {
	hash := strings.Split(repoURL, "@")[1]
	ownerRepo := repoURL[:strings.Index(repoURL, "@")]
	resp, data, errs := gorequest.New().Get("https://github.com/"+ownerRepo+"/archive/"+hash+".zip").
		Set("User-Agent", "bazaar/1.0.0 https://github.com/siyuan-note/bazaar").
		Timeout(30 * time.Second).EndBytes()
	if nil != errs {
		logger.Errorf("get repo zip failed: %s", errs)
		return nil
	}
	if 200 != resp.StatusCode {
		logger.Errorf("get repo zip failed: %s", errs)
		return nil
	}

	logger.Infof("downloaded package [%s], size [%d]", repoURL, len(data))
	return data
}
