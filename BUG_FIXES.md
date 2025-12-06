# Bug Fixes

## [2025-11-29] 修复主窗口覆盖子窗口的问题

### 问题描述
当子窗口(如设置、关于等对话框)激活时,点击主窗口会将主窗口置于子窗口之上,导致子窗口被遮挡。

### 根本原因
主窗口被设置为"Always On Top"(置顶)状态,并且通过定期检查来强制保持这个状态。即使子窗口打开时,主窗口仍然保持TopMost状态。由于两个窗口都是TopMost,当点击主窗口时,它会获得焦点并被提升到Z-order堆栈的顶部,从而覆盖子窗口。

### 修复方案
修改 `ui/gui.go` 中的定期检查逻辑(第192-227行):

1. **显式禁用置顶**: 当检测到任何子窗口存在时,显式调用 `SetWindowAlwaysOnTop(mainHwnd, false)` 来禁用主窗口的置顶状态
2. **提高响应性**: 将检查频率从 1秒 提高到 200毫秒,使窗口状态切换更加流畅
3. **优化逻辑**: 使用 `hasChildWindow` 变量明确区分有无子窗口的情况,使代码逻辑更清晰

### 技术细节
- 修改文件: `ui/gui.go`
- 修改位置: 第192-227行 (定期置顶检查的goroutine)
- 关键改动:
  - 检查间隔: `1 * time.Second` → `200 * time.Millisecond`
  - 新增逻辑: 子窗口存在时显式禁用主窗口置顶
  - 代码结构: 使用更清晰的条件判断逻辑

---

## [2025-11-29] 为所有子窗口添加ESC键关闭功能

### 功能描述
为所有子窗口(设置、关于、分组管理、快捷方式管理等)添加ESC键快速关闭功能,提升用户体验。

### 实现方案
1. **创建辅助函数**: 新建 `ui/escape_key_handler.go` 文件,实现 `setupEscapeKeyCloseWithShortcut` 函数
2. **应用到所有子窗口**: 在每个子窗口的 `Show()` 调用前添加ESC键处理

### 修改的窗口列表
- ✅ 设置窗口 (`SettingsWindow`)
- ✅ 关于窗口 (`AboutWindow`)
- ✅ 新增分组窗口 (`AddGroupWindow`)
- ✅ 编辑分组窗口 (`EditGroupWindow`)
- ✅ 删除分组窗口 (`DeleteGroupWindow`)
- ✅ 编辑分组对话框 (`EditGroupDialogForWindow`)
- ✅ 删除分组对话框 (`DeleteGroupDialogForWindow`)
- ✅ 快捷方式窗口 (`ShortcutWindow`)
- ✅ 删除快捷方式对话框

### 技术细节
- 新增文件: `ui/escape_key_handler.go`
- 修改文件: `ui/gui.go`
- 实现方式: 使用Fyne的 `desktop.CustomShortcut` 和 `Canvas.AddShortcut` API
- 关键代码:
  ```go
  setupEscapeKeyCloseWithShortcut(win, func() {
      // 清理窗口引用
      l.WindowReference = nil
      // 关闭窗口
      win.Close()
  })
  ```

### 用户体验改进
- 用户可以使用ESC键快速关闭任何子窗口
- 符合常见的UI交互习惯
- 提高操作效率

---

## [2025-11-29] 修复托盘图标回调的Fyne线程错误

### 问题描述
在运行时出现大量 "Error in Fyne call thread" 错误信息:
```
*** Error in Fyne call thread, this should have been called in fyne.Do[AndWait] ***
  From: C:/Users/manxi/.gemini/antigravity/scratch/go musetool/Go-MuseTool/ui/gui.go:143
```

### 根本原因
托盘图标的回调函数(OnShow)在非UI线程中执行,但直接调用了Fyne的UI操作(`l.Window.Show()`和`l.Window.RequestFocus()`),违反了Fyne的线程安全规则。

### 修复方案
使用 `fyne.Do()` 包装所有在托盘回调中的UI操作,确保它们在Fyne的UI线程中执行:

```go
InitTrayWithData(l.getMainWindowIconData(), func() {
    // OnShow callback (Left click or Menu Show)
    // 必须在UI线程中调用Fyne的UI操作
    fyne.Do(func() {
        l.Window.Show()
        l.Window.RequestFocus()
    })
}, ...)
```

### 技术细节
- 修改文件: `ui/gui.go`
- 修改位置: 第141-145行 (托盘图标OnShow回调)
- 关键改动: 使用 `fyne.Do()` 包装UI操作
- 影响范围: 仅影响托盘图标点击显示窗口的操作

### 验证方法
- 运行程序后不再出现 "Error in Fyne call thread" 错误
- 托盘图标点击功能正常工作
- 窗口显示和焦点切换正常

---

## [2025-11-29] 修复最小化托盘和退出对话框Bug

### 问题描述
存在两个关键bug:
1. **最小化托盘功能失效**: 点击关闭按钮后,即使选择"最小化到托盘",窗口也会完全关闭
2. **首次退出没有提示**: 第一次点击关闭按钮时,应该显示选择对话框,但直接退出了程序

### 根本原因
使用 `SetOnClosed` 处理窗口关闭事件存在以下问题:
1. **递归调用**: 在回调中调用 `l.Window.Hide()` 或 `w.Close()` 会再次触发 `SetOnClosed`,导致递归调用
2. **事件拦截失败**: `SetOnClosed` 在窗口已经开始关闭流程后才被调用,无法真正拦截关闭操作
3. **状态管理混乱**: 使用 `closeDialogShowing` 标志试图避免重复显示对话框,但逻辑复杂且容易出错

### 修复方案
使用 `SetCloseIntercept` 替代 `SetOnClosed`:

**修改前** (使用 SetOnClosed):
```go
w.SetOnClosed(func() {
    if l.Config.CloseDialogShown {
        if l.Config.MinimizeToTray {
            l.Window.Hide()  // ❌ 会触发递归调用
        } else {
            w.Close()        // ❌ 会触发递归调用
        }
        return
    }
    // 显示对话框...
})
```

**修改后** (使用 SetCloseIntercept):
```go
w.SetCloseIntercept(func() {
    if l.Config.CloseDialogShown {
        if l.Config.MinimizeToTray {
            l.Window.Hide()  // ✅ 只是隐藏窗口,不触发关闭事件
        } else {
            l.App.Quit()     // ✅ 直接退出应用,不触发关闭事件
        }
        return
    }
    // 显示对话框...
    // 用户选择后调用 l.Window.Hide() 或 l.App.Quit()
})
```

### 关键改进

1. **使用 SetCloseIntercept**: 
   - 在窗口关闭**之前**拦截事件,可以完全控制关闭行为
   - 不会触发递归调用

2. **使用 App.Quit() 替代 w.Close()**:
   - `w.Close()` 会触发关闭事件
   - `l.App.Quit()` 直接退出应用,不触发任何窗口事件

3. **移除 closeDialogShowing 标志**:
   - 不再需要复杂的状态管理
   - 逻辑更清晰简单

### 技术细节
- 修改文件: `ui/gui.go`
- 修改位置: 第231-308行 (窗口关闭处理逻辑)
- 关键改动:
  - `SetOnClosed` → `SetCloseIntercept`
  - `w.Close()` → `l.App.Quit()`
  - 移除 `closeDialogShowing` 标志
  - 简化逻辑流程

### 验证方法
1. **首次关闭测试**:
   - 点击关闭按钮
   - 应该显示选择对话框
   - 选择"最小化到托盘"后,窗口隐藏但程序继续运行
   - 选择"退出程序"后,程序完全退出

2. **记住选择测试**:
   - 勾选"记住我的选择"
   - 下次点击关闭按钮时,直接执行之前选择的操作
   - 不再显示对话框

3. **托盘恢复测试**:
   - 最小化到托盘后
   - 点击托盘图标
   - 窗口正常显示

---

## [2025-11-29] 修复设置窗口被主窗口覆盖的问题

### 问题描述
设置窗口与其他子窗口行为不一致:
- 打开设置窗口后,点击主窗口会覆盖设置窗口
- 其他子窗口(关于、分组管理等)则正常,不会被主窗口覆盖

### 根本原因
设置窗口在显示后缺少 `applyWindowStyle()` 调用,导致窗口没有被设置为置顶(TopMost)状态。

**对比分析**:

设置窗口 (有问题):
```go
settingsWin.Show()
// ❌ 缺少 applyWindowStyle 调用
```

其他窗口 (正常):
```go
aboutWin.Show()
// ✅ 窗口显示后应用样式,确保子窗口在主窗口前面
l.applyWindowStyle(language.T().AboutTitle)
```

### 修复方案
为设置窗口添加 `applyWindowStyle()` 调用,与其他窗口保持一致:

```go
settingsWin.Show()
// 窗口显示后应用样式,确保子窗口在主窗口前面
l.applyWindowStyle(language.T().SettingsTitle)
```

### 技术细节
- 修改文件: `ui/gui.go`
- 修改位置: 第1019-1021行 (设置窗口显示逻辑)
- 关键改动: 添加 `l.applyWindowStyle(language.T().SettingsTitle)` 调用
- 影响范围: 仅影响设置窗口

### applyWindowStyle 的作用
该函数会:
1. 设置窗口为 TopMost (总在最上层)
2. 设置标题栏颜色
3. 设置窗口不在任务栏显示

### 验证方法
1. 打开设置窗口
2. 点击主窗口
3. 设置窗口应该保持在主窗口前面,不被覆盖

