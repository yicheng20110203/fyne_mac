package main

import (
    "fyne.io/fyne"
    "fyne.io/fyne/app"
    "fyne.io/fyne/container"
    "fyne.io/fyne/widget"
    aliyunsmsclient "github.com/KenmyZhang/aliyun-communicate"
    "github.com/yicheng20110203/fyne_mac/cfg"
    "github.com/yicheng20110203/fyne_mac/lib"
    "regexp"
    "strings"
    "time"
)

func main() {
    a := app.New()
    a.Settings().SetTheme(&lib.MyTheme{})
    w := a.NewWindow("短信push工具")
    w.Resize(fyne.NewSize(540, 240))

    signEntry := widget.NewEntry()
    signEntry.SetPlaceHolder("短信签名不包含【】")
    tplEntry := widget.NewEntry()
    tplEntry.SetPlaceHolder("内容模版编号需要到阿里云后台查看")
    textArea := widget.NewMultiLineEntry()
    textArea.SetPlaceHolder("多个手机号码用英文逗号(,)分隔")
    resEntry := widget.NewEntry()
    resEntry.Disable()
    var ms []string
    sendBtn := widget.NewButton("发送短信", func() {
        smsSignName := signEntry.Text
        smsCtxTpl := tplEntry.Text
        mobiles := textArea.Text
        if smsSignName == "" {
            resEntry.SetText("短信签名必填")
            return
        }

        if smsCtxTpl == "" {
            resEntry.SetText("短信内容模拟必填")
            return
        }

        if mobiles == "" {
            resEntry.SetText("发送手机号必填")
            return
        }

        ms = strings.Split(mobiles, "\n")
        if len(ms) == 0 {
            resEntry.SetText("发送手机号填写错误")
            return
        }

        kv := make(map[string]struct{})
        realMs := make([]string, 0)
        for _, v := range ms {
            vv := strings.Trim(v, " ")
            if len(vv) > 0 {
                tmp := strings.Split(v, ",")
                if len(tmp) > 0 {
                    for _, tv := range tmp {
                        kv[tv] = struct{}{}
                    }
                }
            }
        }
        for k, _ := range kv {
            realMs = append(realMs, k)
        }

        //发送短信
        err := _sendSms(smsSignName, smsCtxTpl, realMs)
        if err != nil {
            resEntry.SetText("短信发送失败")
            return
        }

        resEntry.SetText("本次短信发送成功，请去阿里云控制台查看短信下行状态。")
        return
    })
    exitBtn := widget.NewButton("退出", func() {
        a.Quit()
    })

    formItems := []*widget.FormItem{
        widget.NewFormItem("短信签名", signEntry),
        widget.NewFormItem("模版编号", tplEntry),
        widget.NewFormItem("手机号码", textArea),
        widget.NewFormItem("发送状态", resEntry),
    }
    form := widget.NewForm(formItems...)
    box := container.NewVBox(
        form,
        sendBtn,
        exitBtn,
    )

    w.SetContent(box)

    w.Show()
    a.Run()
}

// 发送短信
func _sendSms(signName string, tplCode string, mobiles []string) (err error) {
    if len(mobiles) == 0 {
        return nil
    }

    smsClient := aliyunsmsclient.New(cfg.SmsGateway)
    sendHandler := func(mobile string) (err error) {
        _, err = smsClient.Execute(cfg.SmsAccessKeyId, cfg.SmsAccessKeySecret, mobile, signName, tplCode, "{\"name\": \"\"}")
        if err != nil {
            return err
        }

        return nil
    }

    mobileFilterHandle := func(mobile string) bool {
        reg := `^1([0-9]+)\d{8}$`
        rgx := regexp.MustCompile(reg)
        return rgx.MatchString(mobile)
    }

    realMobiles := make([]string, 0)
    for _, v := range mobiles {
        if !mobileFilterHandle(v) {
            continue
        }
        realMobiles = append(realMobiles, v)
    }

    if len(realMobiles) == 0 {
        return nil
    }

    // 最多发送2000条
    for i := 0; i < 100; i++ {
        if i*cfg.EachSendSize > len(realMobiles) {
            break
        }

        st := i * cfg.EachSendSize
        en := (i + 1) * cfg.EachSendSize
        exit := false
        if en >= len(realMobiles) {
            en = len(realMobiles)
            exit = true
        }

        err = sendHandler(strings.Join(realMobiles[st:en], ","))
        if err != nil {
            return err
        }

        if exit {
            break
        }

        time.Sleep(time.Second)
    }

    return nil
}