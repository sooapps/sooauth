(function () {
  "use strict";

  var script = document.currentScript;
  if (!script) return;

  var clientId = script.getAttribute("data-client-id");
  var mode = script.getAttribute("data-mode") || "signin";
  var targetSel = script.getAttribute("data-target") || "#sooauth-widget";
  var issuerAttr = script.getAttribute("data-issuer");

  if (!clientId) {
    console.error("[sooauth] data-client-id is required");
    return;
  }

  var root = document.querySelector(targetSel);
  if (!root) {
    console.error("[sooauth] target not found:", targetSel);
    return;
  }

  function escapeHtml(str) {
    if (str == null) return "";
    return String(str)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#039;");
  }

  function dispatch(name, detail) {
    document.dispatchEvent(new CustomEvent(name, { detail: detail, bubbles: true }));
  }

  function stripQuery(keys) {
    var url = new URL(window.location.href);
    keys.forEach(function (k) { url.searchParams.delete(k); });
    window.history.replaceState({}, "", url.toString());
  }

  function loadConfig(base) {
    return fetch(base + "/v1/widget/config?client_id=" + encodeURIComponent(clientId))
      .then(function (r) { return r.json(); });
  }

  function exchangeCode(base, code) {
    return fetch(base + "/v1/widget/exchange", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ code: code, client_id: clientId }),
    }).then(function (r) { return r.json().then(function (d) { return { ok: r.ok, d: d }; }); });
  }

  function providerLabel(p) {
    return p === "google" ? "Google" : p === "github" ? "GitHub" : p;
  }

  function policyFromCfg(cfg) {
    var p = (cfg && cfg.password_policy) || {};
    return {
      min_length: Math.max(p.min_length || 8, 8),
      require_uppercase: !!p.require_uppercase,
      require_number: !!p.require_number,
      require_special: !!p.require_special,
    };
  }

  function buildRules(password, policy) {
    var rules = [{
      id: "length",
      label: "At least " + policy.min_length + " characters",
      ok: password.length >= policy.min_length,
    }];
    if (policy.require_uppercase) {
      rules.push({ id: "upper", label: "One uppercase letter", ok: /[A-Z]/.test(password) });
    }
    if (policy.require_number) {
      rules.push({ id: "number", label: "One number", ok: /[0-9]/.test(password) });
    }
    if (policy.require_special) {
      rules.push({ id: "special", label: "One special character", ok: /[^A-Za-z0-9]/.test(password) });
    }
    return rules;
  }

  function meetsPolicy(password, policy) {
    return buildRules(password, policy).every(function (r) { return r.ok; });
  }

  function rulesHtml(password, policy, touched) {
    return buildRules(password, policy).map(function (rule) {
      var color = !touched ? "#5D6B70" : (rule.ok ? "#051B23" : "#051B23");
      var icon = rule.ok && touched ? "✓" : "✗";
      return '<li style="display:flex;align-items:center;gap:8px;font-size:13px;color:' + color + ';margin:0">' +
        '<span aria-hidden="true" style="width:14px;text-align:center">' + icon + "</span>" +
        "<span>" + rule.label + "</span></li>";
    }).join("");
  }

  function getIdentifierInfo(authConfig) {
    var allowed = (authConfig && authConfig.allowed_identifiers) || ["email"];
    var hasEmail = allowed.indexOf("email") !== -1;
    var hasUsername = allowed.indexOf("username") !== -1;
    var hasPhone = allowed.indexOf("phone") !== -1;

    if (hasEmail && hasUsername && hasPhone) {
      return { label: "Email, username or phone", placeholder: "Email, username, or phone", type: "text" };
    }
    if (hasEmail && hasUsername) {
      return { label: "Email or username", placeholder: "Email or username", type: "text" };
    }
    if (hasEmail && hasPhone) {
      return { label: "Email or phone", placeholder: "Email or phone number", type: "text" };
    }
    if (hasUsername && hasPhone) {
      return { label: "Username or phone", placeholder: "Username or phone number", type: "text" };
    }
    if (hasUsername) {
      return { label: "Username", placeholder: "Username", type: "text" };
    }
    if (hasPhone) {
      return { label: "Phone number", placeholder: "+1234567890", type: "tel" };
    }
    return { label: "Email", placeholder: "Email", type: "email" };
  }

  var currentMode = mode;
  var lastEmail = "";
  var lastIdentifier = "";

  function render(cfg, base) {
    var accent = cfg.accent_color || "#FF3B3B";
    var accentFg = (accent.toLowerCase() === "#ff3b3b") ? "#FFFFFF" : "#051B23";
    var policy = policyFromCfg(cfg);
    var authCfg = cfg.auth_config || { allowed_identifiers: ["email"], primary_auth_mode: "password" };
    var allowedIds = authCfg.allowed_identifiers || ["email"];
    var isOTPMode = authCfg.primary_auth_mode === "otp";
    var idInfo = getIdentifierInfo(authCfg);

    var html = '<div class="sooauth-widget" style="font-family:system-ui,sans-serif;max-width:400px">';

    if (currentMode === "forgot") {
      html += '<p style="margin:0 0 6px;font-size:16px;font-weight:600;color:#051B23">Reset password</p>' +
        '<p style="margin:0 0 14px;font-size:13px;color:#5D6B70">Enter your email and we\'ll send you instructions.</p>' +
        '<form id="sooauth-forgot-form" style="display:grid;gap:10px">' +
        '<input name="email" type="email" required placeholder="Email" value="' + escapeHtml(lastEmail || lastIdentifier) + '" style="padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23" />' +
        '<button id="sooauth-submit" type="submit" style="padding:12px;border:1px solid ' + accent + ';border-radius:8px;background:' + accent + ';color:' + accentFg + ';font-weight:600;cursor:pointer">Send instructions</button>' +
        '</form>' +
        '<div style="display:flex;justify-content:space-between;align-items:center;margin-top:14px;font-size:13px">' +
        '<button id="sooauth-back-btn" type="button" style="background:none;border:none;padding:0;color:#051B23;cursor:pointer;text-decoration:underline">Back to sign in</button>' +
        '<button id="sooauth-code-btn" type="button" style="background:none;border:none;padding:0;color:#5D6B70;cursor:pointer;text-decoration:underline">Have a 6-digit code?</button>' +
        '</div>' +
        '<p id="sooauth-msg" style="font-size:13px;color:#051B23;margin:10px 0 0"></p>';
    } else if (currentMode === "reset_code") {
      html += '<p style="margin:0 0 6px;font-size:16px;font-weight:600;color:#051B23">Choose new password</p>' +
        '<p style="margin:0 0 14px;font-size:13px;color:#5D6B70">Enter your email, the 6-digit code, and your new password.</p>' +
        '<form id="sooauth-reset-form" style="display:grid;gap:10px">' +
        '<input name="email" type="email" required placeholder="Email" value="' + escapeHtml(lastEmail || lastIdentifier) + '" style="padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23" />' +
        '<input name="code" type="text" inputmode="numeric" pattern="[0-9]{6}" maxlength="6" required placeholder="6-digit reset code" style="padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23;font-family:monospace;letter-spacing:0.2em;text-align:center" />' +
        '<div>' +
        '<input id="sooauth-password" name="password" type="password" required placeholder="New password" style="width:100%;box-sizing:border-box;padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23" />' +
        '<ul id="sooauth-password-rules" style="list-style:none;padding:8px 0 0;margin:0;display:grid;gap:6px"></ul>' +
        '</div>' +
        '<button id="sooauth-submit" type="submit" style="padding:12px;border:1px solid ' + accent + ';border-radius:8px;background:' + accent + ';color:' + accentFg + ';font-weight:600;cursor:pointer">Update password</button>' +
        '</form>' +
        '<div style="margin-top:14px;text-align:center;font-size:13px">' +
        '<button id="sooauth-back-btn" type="button" style="background:none;border:none;padding:0;color:#051B23;cursor:pointer;text-decoration:underline">Back to sign in</button>' +
        '</div>' +
        '<p id="sooauth-msg" style="font-size:13px;color:#051B23;margin:10px 0 0"></p>';
    } else if (currentMode === "otp_verify") {
      html += '<p style="margin:0 0 6px;font-size:16px;font-weight:600;color:#051B23">Enter verification code</p>' +
        '<p style="margin:0 0 14px;font-size:13px;color:#5D6B70">Enter the 6-digit code sent to <strong>' + escapeHtml(lastIdentifier) + '</strong>.</p>' +
        '<form id="sooauth-otp-verify-form" style="display:grid;gap:10px">' +
        '<input name="code" type="text" inputmode="numeric" pattern="[0-9]{6}" maxlength="6" required placeholder="6-digit code" style="padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23;font-family:monospace;letter-spacing:0.2em;text-align:center" />' +
        '<button id="sooauth-submit" type="submit" style="padding:12px;border:1px solid ' + accent + ';border-radius:8px;background:' + accent + ';color:' + accentFg + ';font-weight:600;cursor:pointer">Verify & Sign in</button>' +
        '</form>' +
        '<div style="margin-top:14px;text-align:center;font-size:13px">' +
        '<button id="sooauth-back-btn" type="button" style="background:none;border:none;padding:0;color:#051B23;cursor:pointer;text-decoration:underline">Back to sign in</button>' +
        '</div>' +
        '<p id="sooauth-msg" style="font-size:13px;color:#051B23;margin:10px 0 0"></p>';
    } else {
      if (cfg.providers && cfg.providers.length) {
        cfg.providers.forEach(function (p) {
          var href = base + "/auth/social/" + p + "/start?tenant_id=" + encodeURIComponent(cfg.tenant_id) +
            "&client_id=" + encodeURIComponent(clientId) +
            "&return_to=" + encodeURIComponent(window.location.href);
          html += '<a href="' + href + '" style="display:block;text-align:center;margin:0 0 8px;padding:12px;border:1px solid #051B23;border-radius:8px;text-decoration:none;color:#051B23;font-weight:500">' +
            "Continue with " + providerLabel(p) + "</a>";
        });
        if (cfg.email_password_enabled !== false) {
          html += '<div style="text-align:center;color:#5D6B70;font-size:13px;margin:12px 0">or with ' + escapeHtml(idInfo.label.toLowerCase()) + '</div>';
        }
      }

      if (cfg.email_password_enabled !== false) {
        if (currentMode === "signup") {
          // --- SIGN UP FORM ---
          html += '<form id="sooauth-signup-form" style="display:grid;gap:10px">';
          
          // Render allowed identifier fields
          if (allowedIds.indexOf("email") !== -1) {
            html += '<label style="display:grid;gap:4px;font-size:13px;color:#051B23"><span>Email *</span>' +
              '<input name="email" type="email" required placeholder="you@example.com" value="' + escapeHtml(lastEmail) + '" style="padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23" /></label>';
          }
          if (allowedIds.indexOf("username") !== -1) {
            var reqUser = allowedIds.length === 1;
            html += '<label style="display:grid;gap:4px;font-size:13px;color:#051B23"><span>Username' + (reqUser ? ' *' : '') + '</span>' +
              '<input name="username" type="text" ' + (reqUser ? 'required' : '') + ' placeholder="username" style="padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23" /></label>';
          }
          if (allowedIds.indexOf("phone") !== -1) {
            var reqPhone = allowedIds.length === 1;
            html += '<label style="display:grid;gap:4px;font-size:13px;color:#051B23"><span>Phone number' + (reqPhone ? ' *' : '') + '</span>' +
              '<input name="phone" type="tel" ' + (reqPhone ? 'required' : '') + ' placeholder="+1234567890" style="padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23" /></label>';
          }

          // Password if not OTP mode
          if (!isOTPMode) {
            html += '<div>' +
              '<label style="display:grid;gap:4px;font-size:13px;color:#051B23"><span>Password *</span>' +
              '<input id="sooauth-password" name="password" type="password" required placeholder="Password" style="width:100%;box-sizing:border-box;padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23" /></label>' +
              '<ul id="sooauth-password-rules" style="list-style:none;padding:8px 0 0;margin:0;display:grid;gap:6px"></ul>' +
              '</div>';
          }

          // Dynamic Custom Registration Fields
          if (cfg.registration_schema && cfg.registration_schema.length) {
            cfg.registration_schema.forEach(function (f) {
              var req = f.required ? " *" : "";
              var reqAttr = f.required ? "required" : "";
              var descHtml = f.description ? '<span style="font-size:11px;color:#5D6B70">' + escapeHtml(f.description) + '</span>' : '';

              if (f.type === "textarea") {
                html += '<label style="display:grid;gap:4px;font-size:13px;color:#051B23"><span>' + escapeHtml(f.label) + req + '</span>' +
                  '<textarea name="custom_' + escapeHtml(f.id) + '" placeholder="' + escapeHtml(f.placeholder || '') + '" ' + reqAttr + ' style="padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23;min-height:60px;font-family:inherit;resize:vertical"></textarea>' +
                  descHtml + '</label>';
              } else if (f.type === "number") {
                html += '<label style="display:grid;gap:4px;font-size:13px;color:#051B23"><span>' + escapeHtml(f.label) + req + '</span>' +
                  '<input name="custom_' + escapeHtml(f.id) + '" type="number" placeholder="' + escapeHtml(f.placeholder || '') + '" ' + reqAttr + ' style="padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23" />' +
                  descHtml + '</label>';
              } else if (f.type === "checkbox") {
                html += '<div style="margin:2px 0"><label style="display:flex;align-items:center;gap:8px;font-size:13px;color:#051B23;cursor:pointer">' +
                  '<input type="checkbox" name="custom_' + escapeHtml(f.id) + '" ' + reqAttr + ' style="width:16px;height:16px;margin:0;cursor:pointer" />' +
                  '<span>' + escapeHtml(f.label) + req + '</span></label>' +
                  (f.description ? '<div style="font-size:11px;color:#5D6B70;margin-left:24px">' + escapeHtml(f.description) + '</div>' : '') +
                  '</div>';
              } else if (f.type === "select") {
                var optsHtml = '<option value="">' + (f.placeholder ? escapeHtml(f.placeholder) : (f.options_url ? 'Loading options…' : '-- Select --')) + '</option>';
                if (f.options && f.options.length) {
                  f.options.forEach(function (opt) {
                    optsHtml += '<option value="' + escapeHtml(opt.value) + '">' + escapeHtml(opt.label) + '</option>';
                  });
                }
                var optsUrlAttr = f.options_url ? ' data-options-url="' + escapeHtml(f.options_url) + '"' : '';
                html += '<label style="display:grid;gap:4px;font-size:13px;color:#051B23"><span>' + escapeHtml(f.label) + req + '</span>' +
                  '<select name="custom_' + escapeHtml(f.id) + '" data-field-id="' + escapeHtml(f.id) + '"' + optsUrlAttr + ' ' + reqAttr + ' style="padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23;background:white">' +
                  optsHtml + '</select>' + descHtml + '</label>';
              } else {
                // Default text
                html += '<label style="display:grid;gap:4px;font-size:13px;color:#051B23"><span>' + escapeHtml(f.label) + req + '</span>' +
                  '<input name="custom_' + escapeHtml(f.id) + '" type="text" placeholder="' + escapeHtml(f.placeholder || '') + '" ' + reqAttr + ' style="padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23" />' +
                  descHtml + '</label>';
              }
            });
          }

          html += '<button id="sooauth-submit" type="submit" style="padding:12px;border:1px solid ' + accent + ';border-radius:8px;background:' + accent + ';color:' + accentFg + ';font-weight:600;cursor:pointer">' +
            "Create account</button></form>" +
            '<div style="margin-top:14px;text-align:center;font-size:13px">' +
            '<button id="sooauth-toggle-mode" type="button" style="background:none;border:none;padding:0;color:#5D6B70;cursor:pointer;text-decoration:underline">Already have an account? Sign in</button>' +
            '</div>' +
            '<p id="sooauth-msg" style="font-size:13px;color:#051B23;margin:8px 0 0"></p>';
        } else {
          // --- SIGN IN FORM ---
          if (isOTPMode) {
            html += '<form id="sooauth-otp-send-form" style="display:grid;gap:10px">' +
              '<input name="identifier" type="' + idInfo.type + '" required placeholder="' + escapeHtml(idInfo.placeholder) + '" value="' + escapeHtml(lastIdentifier || lastEmail) + '" style="padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23" />' +
              '<button id="sooauth-submit" type="submit" style="padding:12px;border:1px solid ' + accent + ';border-radius:8px;background:' + accent + ';color:' + accentFg + ';font-weight:600;cursor:pointer">Send verification code</button></form>' +
              '<div style="margin-top:14px;text-align:center;font-size:13px">' +
              '<button id="sooauth-toggle-mode" type="button" style="background:none;border:none;padding:0;color:#5D6B70;cursor:pointer;text-decoration:underline">Don\'t have an account? Sign up</button>' +
              '</div>' +
              '<p id="sooauth-msg" style="font-size:13px;color:#051B23;margin:8px 0 0"></p>';
          } else {
            var showRememberMe = cfg.remember_me_enabled !== false;
            var rememberMeChecked = cfg.remember_me_default !== false;
            var allowForgot = allowedIds.indexOf("email") !== -1;
            html += '<form id="sooauth-email-form" style="display:grid;gap:10px">' +
              '<input name="identifier" type="' + idInfo.type + '" required placeholder="' + escapeHtml(idInfo.placeholder) + '" value="' + escapeHtml(lastIdentifier || lastEmail) + '" style="padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23" />' +
              '<div>' +
              '<input id="sooauth-password" name="password" type="password" required placeholder="Password" style="width:100%;box-sizing:border-box;padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23" />' +
              (allowForgot ? '<div style="display:flex;justify-content:flex-end;margin-top:4px"><button id="sooauth-forgot-btn" type="button" style="background:none;border:none;padding:0;color:#5D6B70;font-size:12px;cursor:pointer;text-decoration:underline">Forgot password?</button></div>' : '') +
              '</div>' +
              (showRememberMe ? '<label style="display:flex;align-items:center;gap:8px;font-size:13px;color:#051B23;cursor:pointer;user-select:none"><input type="checkbox" name="remember_me" value="true"' + (rememberMeChecked ? " checked" : "") + ' style="width:16px;height:16px;margin:0;cursor:pointer" /><span>Remember me</span></label>' : '') +
              '<button id="sooauth-submit" type="submit" style="padding:12px;border:1px solid ' + accent + ';border-radius:8px;background:' + accent + ';color:' + accentFg + ';font-weight:600;cursor:pointer">Sign in</button></form>' +
              '<div style="margin-top:14px;text-align:center;font-size:13px">' +
              '<button id="sooauth-toggle-mode" type="button" style="background:none;border:none;padding:0;color:#5D6B70;cursor:pointer;text-decoration:underline">Don\'t have an account? Sign up</button>' +
              '</div>' +
              '<p id="sooauth-msg" style="font-size:13px;color:#051B23;margin:8px 0 0"></p>';
          }
        }
      }
    }

    if (cfg.show_powered_by) {
      html += '<p style="margin-top:16px;font-size:12px;color:#5D6B70;text-align:center">Powered by <a href="https://sooauth.com" style="color:#051B23">sooauth</a></p>';
    }
    html += "</div>";
    root.innerHTML = html;

    // Fetch dynamic select options if present
    var dynamicSelects = root.querySelectorAll("select[data-options-url]");
    dynamicSelects.forEach(function (sel) {
      var url = sel.getAttribute("data-options-url");
      if (!url) return;
      fetch(url)
        .then(function (r) { return r.json(); })
        .then(function (d) {
          if (d && Array.isArray(d.options)) {
            var placeholder = sel.querySelector("option[value='']") ? sel.querySelector("option[value='']").textContent : "-- Select --";
            if (placeholder === "Loading options…" || placeholder === "-- Select --") placeholder = "-- Select --";
            sel.innerHTML = '<option value="">' + escapeHtml(placeholder) + '</option>' +
              d.options.map(function (opt) {
                return '<option value="' + escapeHtml(opt.value) + '">' + escapeHtml(opt.label) + '</option>';
              }).join("");
          }
        })
        .catch(function (err) {
          console.error("[sooauth] failed to fetch options for " + sel.name, err);
        });
    });

    var backBtn = document.getElementById("sooauth-back-btn");
    if (backBtn) {
      backBtn.onclick = function () {
        currentMode = "signin";
        render(cfg, base);
      };
    }

    var toggleBtn = document.getElementById("sooauth-toggle-mode");
    if (toggleBtn) {
      toggleBtn.onclick = function () {
        currentMode = (currentMode === "signup" ? "signin" : "signup");
        render(cfg, base);
      };
    }

    var codeBtn = document.getElementById("sooauth-code-btn");
    if (codeBtn) {
      codeBtn.onclick = function () {
        currentMode = "reset_code";
        render(cfg, base);
      };
    }

    var forgotBtn = document.getElementById("sooauth-forgot-btn");
    if (forgotBtn) {
      forgotBtn.onclick = function () {
        currentMode = "forgot";
        render(cfg, base);
      };
    }

    // --- FORGOT FORM ---
    var forgotForm = document.getElementById("sooauth-forgot-form");
    if (forgotForm) {
      forgotForm.addEventListener("submit", function (e) {
        e.preventDefault();
        var fd = new FormData(forgotForm);
        var emailVal = String(fd.get("email") || "").trim();
        lastEmail = emailVal;
        lastIdentifier = emailVal;
        var msg = document.getElementById("sooauth-msg");
        msg.style.color = "#051B23";
        msg.textContent = "Sending…";

        fetch(base + "/auth/forgot-password", {
          method: "POST",
          headers: { "Content-Type": "application/json", "X-Sooauth-Embed": "1" },
          body: JSON.stringify({
            email: emailVal,
            client_id: clientId,
            return_to: window.location.href,
          }),
        })
          .then(function (r) { return r.json().then(function (d) { return { ok: r.ok, d: d }; }); })
          .then(function (res) {
            if (!res.ok) {
              msg.textContent = res.d.message || res.d.error || "Request failed";
              return;
            }
            if (res.d && res.d.delivery === "code") {
              msg.textContent = "6-digit reset code sent to your email. Opening verification form…";
              setTimeout(function () {
                currentMode = "reset_code";
                render(cfg, base);
              }, 1200);
            } else {
              msg.textContent = "Reset instructions sent to your email. Check your inbox.";
            }
          })
          .catch(function () {
            msg.textContent = "Network error. Please try again.";
          });
      });
      return;
    }

    // --- RESET FORM ---
    var resetForm = document.getElementById("sooauth-reset-form");
    var passwordInput = document.getElementById("sooauth-password");
    var rulesEl = document.getElementById("sooauth-password-rules");
    var submitBtn = document.getElementById("sooauth-submit");
    var passwordTouched = false;

    function refreshPasswordUi() {
      if (!rulesEl || !passwordInput || (currentMode !== "signup" && currentMode !== "reset_code")) return;
      var val = passwordInput.value || "";
      if (document.activeElement === passwordInput) passwordTouched = true;
      if (val.length > 0) passwordTouched = true;
      rulesEl.innerHTML = rulesHtml(val, policy, passwordTouched);
      if (submitBtn) {
        var ok = meetsPolicy(val, policy);
        submitBtn.disabled = !ok;
        submitBtn.style.opacity = ok ? "1" : "0.5";
        submitBtn.style.cursor = ok ? "pointer" : "not-allowed";
      }
    }

    if (passwordInput && (currentMode === "signup" || currentMode === "reset_code")) {
      passwordInput.addEventListener("input", refreshPasswordUi);
      passwordInput.addEventListener("focus", function () {
        passwordTouched = true;
        refreshPasswordUi();
      });
      refreshPasswordUi();
    }

    if (resetForm) {
      resetForm.addEventListener("submit", function (e) {
        e.preventDefault();
        var fd = new FormData(resetForm);
        var emailVal = String(fd.get("email") || "").trim();
        var codeVal = String(fd.get("code") || "").trim();
        var passVal = String(fd.get("password") || "");
        lastEmail = emailVal;
        lastIdentifier = emailVal;
        var msg = document.getElementById("sooauth-msg");
        msg.style.color = "#051B23";

        if (!meetsPolicy(passVal, policy)) {
          passwordTouched = true;
          refreshPasswordUi();
          return;
        }

        msg.textContent = "Updating password…";
        fetch(base + "/auth/reset-password", {
          method: "POST",
          headers: { "Content-Type": "application/json", "X-Sooauth-Embed": "1" },
          body: JSON.stringify({
            email: emailVal,
            code: codeVal,
            password: passVal,
            client_id: clientId,
          }),
        })
          .then(function (r) { return r.json().then(function (d) { return { ok: r.ok, d: d }; }); })
          .then(function (res) {
            if (!res.ok) {
              msg.textContent = res.d.message || res.d.error || "Password reset failed";
              return;
            }
            msg.textContent = "Password updated! Redirecting to sign in…";
            dispatch("sooauth:password_reset", { email: emailVal });
            setTimeout(function () {
              currentMode = "signin";
              render(cfg, base);
            }, 1200);
          })
          .catch(function () {
            msg.textContent = "Network error. Please try again.";
          });
      });
      return;
    }

    // --- OTP SEND FORM ---
    var otpSendForm = document.getElementById("sooauth-otp-send-form");
    if (otpSendForm) {
      otpSendForm.addEventListener("submit", function (e) {
        e.preventDefault();
        var fd = new FormData(otpSendForm);
        var idVal = String(fd.get("identifier") || "").trim();
        lastIdentifier = idVal;
        var msg = document.getElementById("sooauth-msg");
        msg.style.color = "#051B23";
        msg.textContent = "Sending code…";

        fetch(base + "/auth/otp/send", {
          method: "POST",
          headers: { "Content-Type": "application/json", "X-Sooauth-Embed": "1" },
          body: JSON.stringify({
            client_id: clientId,
            identifier: idVal,
          }),
        })
          .then(function (r) { return r.json().then(function (d) { return { ok: r.ok, d: d }; }); })
          .then(function (res) {
            if (!res.ok) {
              msg.textContent = res.d.message || res.d.error || "Failed to send code";
              return;
            }
            currentMode = "otp_verify";
            render(cfg, base);
          })
          .catch(function () {
            msg.textContent = "Network error. Please try again.";
          });
      });
      return;
    }

    // --- OTP VERIFY FORM ---
    var otpVerifyForm = document.getElementById("sooauth-otp-verify-form");
    if (otpVerifyForm) {
      otpVerifyForm.addEventListener("submit", function (e) {
        e.preventDefault();
        var fd = new FormData(otpVerifyForm);
        var codeVal = String(fd.get("code") || "").trim();
        var msg = document.getElementById("sooauth-msg");
        msg.style.color = "#051B23";
        msg.textContent = "Verifying…";

        fetch(base + "/auth/otp/verify", {
          method: "POST",
          headers: { "Content-Type": "application/json", "X-Sooauth-Embed": "1" },
          body: JSON.stringify({
            client_id: clientId,
            identifier: lastIdentifier,
            code: codeVal,
          }),
        })
          .then(function (r) { return r.json().then(function (d) { return { ok: r.ok, d: d }; }); })
          .then(function (res) {
            if (!res.ok) {
              msg.textContent = res.d.message || res.d.error || "Invalid code";
              return;
            }
            dispatch("sooauth:login", {
              access_token: res.d.access_token,
              email: res.d.email,
              expires_in: res.d.expires_in,
            });
            msg.textContent = "Signed in.";
          })
          .catch(function () {
            msg.textContent = "Network error. Please try again.";
          });
      });
      return;
    }

    // --- SIGN UP FORM ---
    var signupForm = document.getElementById("sooauth-signup-form");
    if (signupForm) {
      signupForm.addEventListener("submit", function (e) {
        e.preventDefault();
        var fd = new FormData(signupForm);
        var emailVal = String(fd.get("email") || "").trim();
        var usernameVal = String(fd.get("username") || "").trim();
        var phoneVal = String(fd.get("phone") || "").trim();
        var passwordVal = String(fd.get("password") || "");
        var msg = document.getElementById("sooauth-msg");
        msg.textContent = "";

        if (!emailVal && !usernameVal && !phoneVal) {
          msg.textContent = "Please provide at least one identifier.";
          return;
        }

        if (!isOTPMode && !meetsPolicy(passwordVal, policy)) {
          passwordTouched = true;
          refreshPasswordUi();
          return;
        }

        // Collect metadata
        var metadata = {};
        if (cfg.registration_schema && cfg.registration_schema.length) {
          cfg.registration_schema.forEach(function (f) {
            var el = signupForm.querySelector('[name="custom_' + f.id + '"]');
            if (el) {
              if (f.type === "checkbox") {
                metadata[f.id] = el.checked;
              } else if (f.type === "number") {
                var n = parseFloat(el.value);
                metadata[f.id] = isNaN(n) ? el.value : n;
              } else {
                metadata[f.id] = el.value;
              }
            }
          });
        }

        var payload = {
          client_id: clientId,
          metadata: metadata,
        };
        if (emailVal) payload.email = emailVal;
        if (usernameVal) payload.username = usernameVal;
        if (phoneVal) payload.phone = phoneVal;
        if (passwordVal) payload.password = passwordVal;

        msg.textContent = "Creating account…";
        fetch(base + "/auth/sign-up", {
          method: "POST",
          headers: { "Content-Type": "application/json", "X-Sooauth-Embed": "1" },
          body: JSON.stringify(payload),
        })
          .then(function (r) { return r.json().then(function (d) { return { ok: r.ok, d: d }; }); })
          .then(function (res) {
            if (!res.ok) {
              msg.textContent = res.d.message || res.d.error || "Registration failed";
              return;
            }
            if (isOTPMode) {
              lastIdentifier = phoneVal || emailVal || usernameVal;
              // Trigger OTP send
              fetch(base + "/auth/otp/send", {
                method: "POST",
                headers: { "Content-Type": "application/json", "X-Sooauth-Embed": "1" },
                body: JSON.stringify({ client_id: clientId, identifier: lastIdentifier }),
              }).finally(function () {
                currentMode = "otp_verify";
                render(cfg, base);
              });
            } else {
              msg.textContent = res.d.message || "Account created! You can now sign in.";
            }
          })
          .catch(function () {
            msg.textContent = "Network error. Please try again.";
          });
      });
      return;
    }

    // --- STANDARD SIGN IN FORM ---
    var form = document.getElementById("sooauth-email-form");
    if (!form) return;

    form.addEventListener("submit", function (e) {
      e.preventDefault();
      var fd = new FormData(form);
      var password = String(fd.get("password") || "");
      var idVal = String(fd.get("identifier") || "").trim();
      lastIdentifier = idVal;
      lastEmail = idVal;
      var msg = document.getElementById("sooauth-msg");
      msg.textContent = "";

      var payload = {
        identifier: idVal,
        email: idVal,
        password: password,
        client_id: clientId,
      };
      var rememberEl = form.querySelector('input[name="remember_me"]');
      if (rememberEl) {
        payload.remember_me = rememberEl.checked;
      }
      fetch(base + "/auth/sign-in", {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-Sooauth-Embed": "1" },
        body: JSON.stringify(payload),
      })
        .then(function (r) { return r.json().then(function (d) { return { ok: r.ok, d: d }; }); })
        .then(function (res) {
          if (!res.ok) {
            msg.style.color = "#051B23";
            msg.textContent = res.d.message || res.d.error || "Request failed";
            return;
          }
          dispatch("sooauth:login", {
            access_token: res.d.access_token,
            email: res.d.email,
            expires_in: res.d.expires_in,
          });
          msg.style.color = "#051B23";
          msg.textContent = "Signed in.";
        })
        .catch(function () {
          msg.textContent = "Network error. Please try again.";
        });
    });
  }

  function boot(base) {
    loadConfig(base).then(function (cfg) {
      var params = new URLSearchParams(window.location.search);
      var widgetCode = params.get("widget_code");
      var err = params.get("error");
      if (err) {
        root.innerHTML = '<p style="font:14px system-ui;color:#051B23">Sign-in failed: ' + escapeHtml(err) + "</p>";
        stripQuery(["error", "widget_code"]);
        return;
      }
      if (widgetCode) {
        root.innerHTML = '<p style="font:14px system-ui;color:#5D6B70">Completing sign-in…</p>';
        exchangeCode(base, widgetCode).then(function (res) {
          stripQuery(["widget_code"]);
          if (!res.ok) {
            root.innerHTML = '<p style="font:14px system-ui;color:#051B23">Could not complete sign-in.</p>';
            return;
          }
          dispatch("sooauth:login", res.d);
          root.innerHTML = '<p style="font:14px system-ui;color:#051B23">Signed in as ' + escapeHtml(res.d.email || "user") + ".</p>";
        });
        return;
      }
      render(cfg, base);
    }).catch(function () {
      root.innerHTML = "<p style=\"font:14px system-ui;color:#051B23\">Could not load sooauth widget</p>";
    });
  }

  var baseGuess = issuerAttr || (script.src ? new URL(script.src).origin : "");
  boot(baseGuess.replace(/\/$/, ""));
})();
