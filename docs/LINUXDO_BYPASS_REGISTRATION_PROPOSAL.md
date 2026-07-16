# LinuxDo `bypass_registration` 改造方案

> 目标：为 LinuxDo Connect 增加 `linuxdo_connect_bypass_registration`，**完全对齐钉钉**「关开放注册仍允许某渠道建号」。  
> 依据：钉钉现网实现 + [AUTH_REGISTRATION_LINUXDO_ANALYSIS.md](./AUTH_REGISTRATION_LINUXDO_ANALYSIS.md) + 注册旁路难度分析。  
> 关联方案：[LINUXDO_EMAIL_BYPASS_PROPOSAL.md](./LINUXDO_EMAIL_BYPASS_PROPOSAL.md)（邮箱关联豁免，与本方案正交）。  
> 范围：设计方案；**已落地实施（见仓库代码）**。  
> 日期：2026-07-15
>
> **实施状态（2026-07-15）**：linuxdo_connect_bypass_registration 配置面 + service canBypass + handler signupBlocked 已合并；默认 false；无 schema migration。

---

## 1. 背景与问题

### 1.1 运维诉求

希望配置为：

| 开关 | 期望 |
|------|------|
| 开放注册 `registration_enabled` | **OFF**（关闭邮箱密码自助注册） |
| LinuxDo Connect | **ON** |
| **新增** `linuxdo_connect_bypass_registration` | **ON** |

效果：

- 邮箱 `/auth/register`：**仍禁止**
- 已有用户 / 已绑定 LinuxDo identity：**可登录**
- **全新** LinuxDo 用户：仍允许建号（与钉钉「开放钉钉注册」同语义）

### 1.2 现状

| 能力 | 钉钉 | LinuxDo |
|------|------|---------|
| 配置开关 | `dingtalk_connect_bypass_registration` | **无** |
| Service 旁路 | `canBypassRegistrationDisabledForOAuth` 仅 `dingtalk` | 恒 `false` |
| 信任边界 | 必须 `corp_restriction_policy=internal_only` | 无企业 corp 模型 |
| Handler 提前分流 | `isDingTalkSignupBlocked` → `bind_login_required` | **无** `signupBlocked`；`create_account_allowed` 写死 `true` |

关闭开放注册后，LinuxDo 新用户在 service 层直接 `ErrRegDisabled`（`loginOrRegisterOAuthWithTokenPair` / 补邮箱建号路径）。

### 1.3 钉钉权威语义（必须对齐）

钉钉管理端文案：

- 标题：`开放钉钉注册`
- Hint：`即使「开放注册」关闭时也可以通过钉钉登录来注册`

技术条件（**两者同时满足**才旁路）：

```text
cfg.Enabled
AND cfg.BypassRegistration
AND cfg.CorpRestrictionPolicy == "internal_only"
```

UI 上该 Toggle **仅在** `corp_restriction_policy === 'internal_only'` 时展示；保存时若策略不是 internal_only，后端会强制 `bypass_registration=false`。

Service：

```go
func (s *AuthService) canBypassRegistrationDisabledForOAuth(ctx context.Context, signupSource string) bool {
    if signupSource != "dingtalk" { return false }
    cfg, err := s.settingService.GetDingTalkConnectOAuthConfig(ctx)
    if err != nil || !cfg.Enabled || !cfg.BypassRegistration { return false }
    return cfg.CorpRestrictionPolicy == "internal_only"
}
```

Handler（callback 早期）：

```go
// 注册关且未豁免 → signupBlocked=true
// → create_account_allowed=false，必要时 step=bind_login_required
// 避免用户填完补邮箱表单才收到 REGISTRATION_DISABLED
```

---

## 2. 目标与非目标

### 2.1 目标

1. 新增 `linuxdo_connect_bypass_registration`，语义与钉钉「开放 xx 注册」一致：  
   **`registration_enabled=false` 时，仍允许经 LinuxDo OAuth 创建本地用户**。
2. Service 层统一走扩展后的 `canBypassRegistrationDisabledForOAuth`，覆盖静默建号、complete-registration、补邮箱建号。
3. Handler 层补齐 `isLinuxDoSignupBlocked` + pending payload（`create_account_allowed` / `bind_login_required`），对齐钉钉 UX。
4. 默认 **关闭**；升级行为不变。
5. **不**绕过 Backend Mode、邀请码、邮箱验证 / force_email（与钉钉一致：旁路只针对 `registration_enabled`）。

### 2.2 非目标

- 不修改邮箱密码注册门禁。
- 不把旁路扩散到微信 / OIDC / GitHub / Google（可二期通用化）。
- 不替代 [LINUXDO_EMAIL_BYPASS_PROPOSAL.md](./LINUXDO_EMAIL_BYPASS_PROPOSAL.md) 的邮箱关联豁免。
- 不在本方案中实现 LinuxDo trust level / 用户组 allowlist（列为强烈建议的**可选**加固，见 §4.3）。

### 2.3 「完全对齐」的含义

| 对齐项 | 钉钉 | LinuxDo 本方案 |
|--------|------|----------------|
| Setting 命名 | `*_bypass_registration` | `linuxdo_connect_bypass_registration` |
| 文案结构 | 「开放钉钉注册」 | 「开放 LinuxDo 注册」 |
| Service API | `canBypassRegistrationDisabledForOAuth` | **同一函数**扩展分支 |
| 建号门表达式 | `IsRegistrationEnabled \|\| canBypass(...)` | 不变，仅扩展 canBypass |
| Callback 提前分流 | `signupBlocked` | 镜像实现 |
| 默认值 | false | false |
| 与 Backend Mode | 不旁路 | 不旁路 |
| 信任边界 | `internal_only` | **无 corp**：用「启用 + flag」对齐**开关形态**；用文档/可选邀请码约束补**风险**（§4） |

> 说明：钉钉的 `internal_only` 是**企业身份硬约束**，LinuxDo 协议层没有等价字段。  
> 「完全对齐」指 **产品开关 + 代码路径 + UX 分流 + 默认安全** 对齐；信任模型差异必须在 UI Hint 与运维文档中显式写出，避免运维误以为「和钉钉一样安全」。

---

## 3. 现状卡点清单（改造对照表）

| # | 位置 | 现状 | 本方案动作 |
|---|------|------|------------|
| 1 | `canBypassRegistrationDisabledForOAuth` | 仅 dingtalk | **扩展 linuxdo 分支** |
| 2 | `loginOrRegisterOAuthWithTokenPair` | 用 canBypass | 随 #1 自动生效 |
| 3 | `RegisterOAuthEmailAccount` / `RegisterVerifiedOAuthEmailAccount` | 用 canBypass | 随 #1 |
| 4 | `CompleteLinuxDoOAuthRegistration` | 调 #2 | 随 #1 |
| 5 | `LinuxDoOAuthCallback` 静默建号 | 调 #2；失败当 session_error | 随 #1；错误体验可选优化 |
| 6 | `createLinuxDoOAuthChoicePendingSession` | `create_account_allowed: true` 写死 | **必须改**：接入 signupBlocked |
| 7 | LinuxDo 无 `isLinuxDoSignupBlocked` | 缺 | **新增**（对齐钉钉） |
| 8 | 补邮箱 pending create-account | `ensureBackendMode` + RegisterOAuthEmail | 随 #1；Backend Mode 仍拦 |
| 9 | `createEmailOAuthUser` | **硬编码**只看 IsRegistrationEnabled | LinuxDo 主路径不用；**可不改**（或顺手统一，非必须） |
| 10 | 邮箱 `RegisterWithVerification` | 只看开放注册 | **禁止**被本 flag 影响 |
| 11 | 前端 `LinuxDoCallbackView` | 已有 create/bind 结构 | 优先靠后端 payload；必要时认 `create_account_allowed` |
| 12 | 配置 / Admin UI / i18n / audit | 无字段 | **新增** |

`ensureBackendModeAllowsNewUserLogin`：**不是**开放注册门，本方案 **绝不** 因 bypass_registration 而放行 Backend Mode。

---

## 4. 推荐设计

### 4.1 配置模型

#### Setting Key

```text
linuxdo_connect_bypass_registration   // bool, 默认 false
```

#### Config

```go
// LinuxDoConnectConfig
BypassRegistration bool `mapstructure:"bypass_registration"`
```

`deploy/config.example.yaml`：

```yaml
linuxdo_connect:
  enabled: false
  # ...
  # 开放 LinuxDo 注册：即使全局「开放注册」关闭，仍允许通过 LinuxDo 登录创建账号。
  # 默认 false。LinuxDo 无企业 corp 限制，开启前请评估风险；建议同时开启「邀请码注册」。
  bypass_registration: false
```

生效优先级：与现有 LinuxDo 字段一致 —— **DB 覆盖 YAML**（`GetLinuxDoConnectOAuthConfig`）。

#### 管理端

- DTO / update / parse / audit 增加字段。
- `SettingsView` LinuxDo 卡片：在 `linuxdo_connect_enabled` 时展示 Toggle。
- i18n（对齐钉钉语气）：

| key | zh 建议 |
|-----|---------|
| `bypassRegistration` | 开放 LinuxDo 注册 |
| `bypassRegistrationHint` | 即使「开放注册」关闭时也可以通过 LinuxDo 登录来注册。注意：LinuxDo 无企业成员限制，开启后任意成功完成 OAuth 的 LinuxDo 用户均可建号；建议同时开启邀请码。 |

### 4.2 Service：扩展 `canBypassRegistrationDisabledForOAuth`

```go
func (s *AuthService) canBypassRegistrationDisabledForOAuth(ctx context.Context, signupSource string) bool {
    switch signupSource {
    case "dingtalk":
        cfg, err := s.settingService.GetDingTalkConnectOAuthConfig(ctx)
        if err != nil || !cfg.Enabled || !cfg.BypassRegistration {
            return false
        }
        return cfg.CorpRestrictionPolicy == "internal_only"

    case "linuxdo":
        cfg, err := s.settingService.GetLinuxDoConnectOAuthConfig(ctx)
        if err != nil || !cfg.Enabled || !cfg.BypassRegistration {
            return false
        }
        // LinuxDo 无 internal_only 等价物：Enabled + BypassRegistration 即旁路。
        // 安全默认：配置读取失败 → false（已由 err 分支覆盖）。
        return true

    default:
        return false
    }
}
```

建号门（**保持现有表达式，不改调用方结构**）：

```text
allowNewUser =
  IsRegistrationEnabled(ctx)
  OR canBypassRegistrationDisabledForOAuth(ctx, signupSource)
```

已调用该表达式的路径（linuxdo 自动受益）：

- `loginOrRegisterOAuthWithTokenPair`（`signupSource="linuxdo"`）
- `RegisterOAuthEmailAccount` / `RegisterVerifiedOAuthEmailAccount`

### 4.3 信任边界：与钉钉的差异与补强

| | 钉钉 | LinuxDo（本方案默认） |
|--|------|----------------------|
| 硬约束 | `internal_only` + 企业 corp 校验 | **无**（OAuth subject 即互联网用户） |
| 风险 | 限本企业员工 | 开 bypass ≈「隐式对全 LinuxDo 开放注册」 |

**推荐产品策略（实现时三选一，写入设置校验）：**

| 策略 | 强度 | 实现成本 | 建议 |
|------|------|----------|------|
| **P0 仅 flag** | 弱 | 最低 | 可做 MVP，Hint 强警告 |
| **P1 flag + 强制邀请码** | 中 | 低 | **推荐默认产品策略**：`bypass_registration=true` 保存时要求 `invitation_code_enabled=true`，否则拒绝或自动提示 |
| **P2 flag + trust level allowlist** | 强 | 中高 | 二期：解析 LinuxDo userinfo trust_level 等 |

本方案正文按 **P0 功能对齐钉钉路径** 写全；**P1 作为强烈建议的保存校验**，在 §5.1 给出伪代码，避免「完全对齐开关」却毫无运营刹车。

### 4.4 Handler：对齐 `signupBlocked` UX

#### 4.4.1 新增

```go
// isLinuxDoSignupBlocked：注册总开关关闭且未开启 LinuxDo 注册旁路时返回 true。
// 镜像 isDingTalkSignupBlocked / canBypassRegistrationDisabledForOAuth(linuxdo)。
func (h *AuthHandler) isLinuxDoSignupBlocked(ctx context.Context) bool {
    if h.settingSvc == nil {
        return false // 或 true：更 fail-closed；建议与钉钉一致：settingSvc nil 时 false 仅因钉钉现状，
                     // 更稳妥为 true。实现时二选一，单测钉死。推荐 fail-closed: true
    }
    if h.settingSvc.IsRegistrationEnabled(ctx) {
        return false
    }
    cfg, err := h.settingSvc.GetLinuxDoConnectOAuthConfig(ctx)
    if err != nil || !cfg.Enabled || !cfg.BypassRegistration {
        return true
    }
    return false
}
```

> 钉钉 `isDingTalkSignupBlocked` 在 `settingSvc==nil` 时返回 false；LinuxDo 建议 **fail-closed（true）**，并在测试中明确。若追求「逐行镜像」可保持 false，但文档标注差异。

#### 4.4.2 改造 `createLinuxDoOAuthChoicePendingSession`

增加参数 `signupBlocked bool`（与钉钉 choice session 一致）：

```text
create_account_allowed = !signupBlocked

if signupBlocked {
  // 无 compat 可绑时：step=bind_login_required
  // 有 compat 命中：仍可 bind 已有账号
  // 禁止引导 create_account / 补邮箱建号（用户填完只会失败）
}
```

可复用或抽取与钉钉相同的：

```go
func bindLoginCompletionResponse(redirectTo string) map[string]any {
    return map[string]any{
        "step":                      "bind_login_required",
        "existing_account_bindable": true,
        "create_account_allowed":    false,
        "redirect":                  redirectTo,
    }
}
```

#### 4.4.3 Callback 决策（与钉钉同构）

```text
LinuxDo callback（login intent，无已有 identity）
  signupBlocked = isLinuxDoSignupBlocked(ctx)
  email 分流（邮箱验证 / force_email / bypass_email）照旧

  if 可走静默合成邮箱路径:
      if signupBlocked:
          → pending bind_login_required（不要 LoginOrRegister）
      else:
          → LoginOrRegister...（此时 canBypass 可在注册关时放行）
          → 邀请码不足 → invitation_required pending

  if 进入 choice / create_account pending:
      → 带上 signupBlocked
      → create_account_allowed = !signupBlocked
```

`CompleteLinuxDoOAuthRegistration` / pending create-account：

- 保持 `ensureBackendModeAllowsNewUserLogin`
- 建号依赖 service canBypass（无需在 handler 再写一套注册判断，避免双源）
- 可选：signupBlocked 时 complete-registration 直接 403，防绕过前端

### 4.5 与其它开关的优先级

```text
1. Backend Mode ON
     → 禁止新用户登录/建号（本 bypass 无效）
2. 已有 AuthIdentity
     → 登录（不看开放注册 / bypass）
3. compat_email 命中
     → 绑定流（不建第二账号）
4. registration_enabled ON
     → 正常允许建号（bypass 无意义但无害）
5. registration_enabled OFF
     → bypass_registration ON 且 LinuxDo enabled → 允许 LinuxDo 建号
     → 否则 signupBlocked，仅绑定
6. invitation_code_enabled
     → 仍要有效邀请码（与钉钉相同：旁路不管邀请码）
7. email_verify / force_email / email-bypass
     → 只影响是否补邮箱，不改变 registration 旁路本身
```

### 4.6 组合矩阵

前提：LinuxDo 启用；全新用户；无 identity；Backend Mode OFF。

| 开放注册 | bypass_registration | 邮箱验证 | 期望 |
|---------|---------------------|----------|------|
| ON | * | OFF | 静默建号（现网） |
| ON | * | ON | 补邮箱或静默（视 email-bypass 方案） |
| **OFF** | **OFF** | * | **不可建号**；bind_login / 已有账号 |
| **OFF** | **ON** | OFF | **可静默建号** |
| **OFF** | **ON** | ON | **可建号**但走补邮箱（除非另开 email-bypass） |
| OFF | ON | * + 邀请码 ON | 建号前仍要邀请码 |
| OFF | ON | * + Backend Mode ON | **仍禁止**新用户 |

邮箱密码注册：任意组合下，只要 `registration_enabled=OFF` → 失败。

---

## 5. 代码改动清单

### 5.1 后端

| 文件 | 改动 |
|------|------|
| `backend/internal/config/config.go` | `LinuxDoConnectConfig.BypassRegistration` |
| `backend/internal/service/domain_constants.go` | `SettingKeyLinuxDoConnectBypassRegistration` |
| `backend/internal/service/setting_oauth.go` | `GetLinuxDoConnectOAuthConfig` 合并字段 |
| `backend/internal/service/setting_parse.go` / `setting_update.go` / `settings_view.go` | 读写 |
| `backend/internal/service/auth_service.go` | 扩展 `canBypassRegistrationDisabledForOAuth` |
| `backend/internal/handler/dto/settings.go` | 响应字段 |
| `backend/internal/handler/admin/setting_handler*.go` | update / audit / 响应；可选 P1 校验 |
| `backend/internal/handler/auth_linuxdo_oauth.go` | `isLinuxDoSignupBlocked`、choice session、callback 分流 |
| `deploy/config.example.yaml` | 示例 + 风险注释 |
| 测试 | §6 |

**P1 保存校验（推荐）伪代码：**

```go
if req.LinuxDoConnectBypassRegistration && !req.InvitationCodeEnabled {
    // 选项 A：硬拒绝
    // return BadRequest("LINUXDO_BYPASS_REQUIRES_INVITATION",
    //   "开启「开放 LinuxDo 注册」前请先开启「邀请码注册」")
    // 选项 B：仅 warning 日志 + UI 确认（弱）
}
```

### 5.2 前端

| 文件 | 改动 |
|------|------|
| `frontend/src/api/admin/settings.ts` | 类型 |
| `frontend/src/views/admin/SettingsView.vue` | LinuxDo 卡片 Toggle + 警告文案 |
| `frontend/src/i18n/locales/zh|en/admin/settings.ts` | 文案 |
| `LinuxDoCallbackView.vue` | 若后端返回 `create_account_allowed=false` / `bind_login_required`，确认 UI 隐藏「创建账号」（钉钉已支持则对照修补） |

### 5.3 明确不改

- `RegisterWithVerification` / `SendVerifyCode` 注册策略  
- Backend Mode 中间件与 `ensureBackendModeAllowsNewUserLogin`  
- 钉钉分支逻辑（除共享函数签名注释更新）  
- 邀请码生成与消费模型  

---

## 6. 测试计划

### 6.1 单测

| 用例 | 期望 |
|------|------|
| `canBypass` linuxdo：enabled+bypass | true |
| `canBypass` linuxdo：bypass false / disabled | false |
| `canBypass` dingtalk 原矩阵 | 回归不变 |
| `canBypass` wechat/oidc | false |
| 注册 OFF + bypass ON → LoginOrRegister linuxdo 建号成功 | 合成邮箱用户存在 |
| 注册 OFF + bypass OFF → ErrRegDisabled | |
| 注册 OFF + bypass ON → 邮箱密码 Register 仍失败 | |
| CompleteLinuxDoOAuthRegistration 同上 | |
| RegisterOAuthEmailAccount signupSource=linuxdo 同上 | |
| `isLinuxDoSignupBlocked` 四象限 | |
| choice session：`create_account_allowed` 随 blocked 变化 | |
| 注册 OFF + bypass OFF + 无 compat → bind_login_required | |
| Backend Mode ON + bypass ON → 仍拒绝新用户 | |
| 邀请码 ON + 注册 OFF + bypass ON 无码 → invitation_required | |
| 设置 API round-trip + audit 含新字段 | |
| （若 P1）bypass ON 且 invitation OFF → 保存失败 | |

建议文件：

- `auth_service_register_test.go`（扩展 `TestCanBypassRegistrationDisabledForOAuth`）
- `auth_linuxdo_oauth_test.go`
- admin setting linuxdo round-trip 测试

### 6.2 手工验收

1. 仅关开放注册，bypass OFF：LinuxDo 新号无法建；老号可登录。  
2. 打开 bypass：LinuxDo 新号可建；`/register` 仍 403。  
3. Backend Mode：仍仅管理员。  
4. 邀请码：仍生效。  
5. 邮箱验证 ON：建号形态（补邮箱 vs 静默）符合 email 相关设置，不被本 flag 错误短路。  

---

## 7. 与邮箱 bypass 方案的关系

两套开关正交：

| Flag | 旁路对象 |
|------|----------|
| `linuxdo_connect_bypass_registration` | 全局 **开放注册** |
| `linuxdo_connect_bypass_email_on_signup`（另一文档） | 全局 **邮箱验证** 导致的补邮箱 |

典型组合：

```text
registration=OFF + bypass_registration=ON + email_verify=ON
  → 允许 LinuxDo 建号，但仍可能要求补邮箱（除非再开 email-bypass）

registration=OFF + bypass_registration=ON + email_verify=OFF
  → 最接近「关全站注册，只准 LinuxDo 一键进」
```

实施顺序建议：

1. 先做 **本方案（registration bypass）** —— 接线最贴钉钉、可独立交付。  
2. 再做 email bypass —— 解决「开放注册 ON + 邮箱验证 ON」的静默体验。  
3. 两者都做完后补交叉矩阵测试。

---

## 8. 迁移与兼容

| 项 | 策略 |
|----|------|
| 已有部署 | 默认 false，行为不变 |
| Public settings | **不**暴露 bypass（避免客户端依赖运维开关） |
| 配置热更新 | 与现有 setting 一致 |
| 文档 | 更新关联分析文；运维手册写清风险与邀请码建议 |

---

## 9. 实施步骤与工作量

| 步骤 | 内容 | 预估 |
|------|------|------|
| 1 | Config + setting key + admin API + UI + i18n | 0.25–0.5 d |
| 2 | 扩展 `canBypass` + 单测 | 0.25 d |
| 3 | LinuxDo signupBlocked + choice/callback | 0.5–1 d |
| 4 | 交叉测试（邀请码 / 邮箱验证 / backend mode） | 0.5 d |
| 5 | （可选 P1）保存校验邀请码 | 0.25 d |
| 6 | 文档与发布说明 | 0.25 d |

**合计：约 2～3 人日（含测试）；P0 最小可通约 1～1.5 人日。**

难度结论（复述）：

- **接线难度：低～中**（钉钉样板完整）。  
- **产品风险：中～高**（无 `internal_only`）。  
- 推荐上线捆绑 **邀请码** 或等价约束。

---

## 10. 风险与决策记录

| 风险 | 缓解 |
|------|------|
| 误开 ≈ 对全 LinuxDo 隐式开放注册 | 默认 false；Hint 警告；推荐 P1 强制邀请码 |
| Handler 不补 signupBlocked | 用户白填表单后失败 → 必须做 §4.4 |
| 与 Backend Mode 混淆 | 文档 + 测试钉死不旁路 |
| 双源注册判断 | 只扩展 canBypass，handler 不复制第二套规则 |
| createEmailOAuthUser 未旁路 | LinuxDo 主路径不用；文档注明；防未来误用 |

**实现前确认：**

1. 是否采用 **P1（bypass 依赖邀请码）**？ → 推荐 **是**。  
2. `settingSvc==nil` 时 signupBlocked 是否 fail-closed？ → 推荐 **是**。  
3. 是否与 email-bypass 同 PR？ → 推荐 **分 PR**，先本方案。

---

## 11. 总结

- 钉钉「关开放注册仍允许渠道建号」= `bypass_registration` +（钉钉专属）`internal_only` + service `canBypass` + handler `signupBlocked`。  
- LinuxDo 对齐做法：新增 **`linuxdo_connect_bypass_registration`**，扩展 **同一** `canBypassRegistrationDisabledForOAuth`，并补齐 **signupBlocked / create_account_allowed** 全链路。  
- 无 corp 硬边界是**固有差异**，应用文案、可选邀请码校验与运维文档补齐，而不是假装与钉钉风险等价。  
- 默认关闭；不旁路 Backend Mode、邀请码、邮箱验证；不影响邮箱密码注册。

---

## 12. 相关索引

| 主题 | 位置 |
|------|------|
| 钉钉 canBypass | `backend/internal/service/auth_service.go` |
| 钉钉 signupBlocked | `backend/internal/handler/auth_dingtalk_oauth.go` |
| LinuxDo callback | `backend/internal/handler/auth_linuxdo_oauth.go` |
| 补邮箱建号门 | `backend/internal/service/auth_oauth_email_flow.go` |
| 钉钉 UI / 文案 | `SettingsView.vue` / `zh/admin/settings.ts` → `dingtalk.bypassRegistration*` |
| 注册关联分析 | `docs/AUTH_REGISTRATION_LINUXDO_ANALYSIS.md` |
| 邮箱豁免方案 | `docs/LINUXDO_EMAIL_BYPASS_PROPOSAL.md` |

---

*本文为 `linuxdo_connect_bypass_registration` 设计方案；确认 §10 决策后可按 §9 实施。*
