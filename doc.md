读取调用号

`ausyscall --dump | grep process_vm_writev`

1. 查看设备信息

`getprop`

获取所有系统属性（如型号、Android 版本、硬件信息）

`getprop ro.build.version.release`

仅查看系统版本号

3. 屏幕截图

`screencap /sdcard/screen.png`

录屏：

`screenrecord /sdcard/demo.mp4`

4. 查看电池状态

`dumpsys battery`

强制设置状态：

`dumpsys battery set level 50`
`dumpsys battery set status 2  # 2 = Charging`

恢复默认模拟：

`dumpsys battery reset`

👀 趣味/隐藏命令
1. 彩蛋

`am start -a android.intent.action.MAIN -n com.android.systemui/.egg.EasterEgg`

5. 打开应用或网页

`am start -a android.intent.action.VIEW -d http://www.example.com`

打开某个 App 的 Activity：

`am start -n com.package.name/.ActivityName`