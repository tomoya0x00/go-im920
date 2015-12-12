[![GoDoc](https://godoc.org/github.com/tomoya0x00/go-im920?status.svg)](http://godoc.org/github.com/tomoya0x00/go-im920)
[![Build Status](https://travis-ci.org/tomoya0x00/go-im920.svg?branch=master)](https://travis-ci.org/tomoya0x00/go-im920)
[![Coverage Status](https://coveralls.io/repos/tomoya0x00/go-im920/badge.svg?branch=master&service=github)](https://coveralls.io/github/tomoya0x00/go-im920?branch=master)

# go-im920

A Go library to control the interplan IM920

インタープラン株式会社様の920MHz無線モジュール、IM920用の制御ライブラリです。

## Tested Environment

* ThinkPad T440s(Windows 10) + IM315-USB-RX
* （近日中に確認予定）Raspberry Pi 2 Model B

## Status

* α版です。API変更の可能性があります。
* 一通りのコマンドに対応しています。

## Usage

モジュール間の送受信には、あらかじめ通信対象とする送信モジュールのIDを登録する必要があります。  
登録方法は [取扱説明書（ハードウェア編）](http://www.interplan.co.jp/support/solution/IM315/manual/IM920_HW_manual.pdf)
の「７－７．送信モジュールIDの登録と消去」に記載があります。  
また、「（２）SRIDコマンドによるIDの登録 」は本ライブラリの im920.AddRcvId() でも可能です。

基本的な使用方法は [こちら](https://github.com/tomoya0x00/go-im920/blob/master/example_interface_test.go) をご参照下さい。
データ受信は [こちら](https://github.com/tomoya0x00/go-im920/blob/master/example_test.go) をご参照下さい。

## References

* [920MHz 無線モジュール IM920 取扱説明書（ハードウェア編）](http://www.interplan.co.jp/support/solution/IM315/manual/IM920_HW_manual.pdf)
* [920MHz 無線モジュール（送受信用）IM920 / IM920XT / IM920c 取扱説明書（ソフトウェア編） ](http://www.interplan.co.jp/support/solution/IM315/manual/IM920_SW_manual.pdf)
* [無線モジュール・アプリケーションノート パソコンによる設定・通信の概要 ](http://www.interplan.co.jp/support/solution/IM315/app_note/AN01.pdf)
