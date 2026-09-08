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

  function render(cfg, base) {
    var accent = cfg.accent_color || "#E8FF3F";
    var policy = policyFromCfg(cfg);
    var html = '<div class="sooauth-widget" style="font-family:system-ui,sans-serif;max-width:400px">';
    if (cfg.providers && cfg.providers.length) {
      cfg.providers.forEach(function (p) {
        var href = base + "/auth/social/" + p + "/start?tenant_id=" + encodeURIComponent(cfg.tenant_id) +
          "&client_id=" + encodeURIComponent(clientId) +
          "&return_to=" + encodeURIComponent(window.location.href);
        html += '<a href="' + href + '" style="display:block;text-align:center;margin:0 0 8px;padding:12px;border:1px solid #051B23;border-radius:8px;text-decoration:none;color:#051B23;font-weight:500">' +
          "Continue with " + providerLabel(p) + "</a>";
      });
      if (cfg.email_password_enabled !== false) {
         html += '<div style="text-align:center;color:#5D6B70;font-size:13px;margin:12px 0">or email</div>';
      }
    }
    if (cfg.email_password_enabled !== false) {
      html += '<form id="sooauth-email-form" style="display:grid;gap:10px">' +
         '<input name="email" type="email" required placeholder="Email" style="padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23" />' +
        '<div>' +
         '<input id="sooauth-password" name="password" type="password" required placeholder="Password" style="width:100%;box-sizing:border-box;padding:10px;border:1px solid #C4C4C4;border-radius:8px;color:#051B23" />' +
        (mode === "signup"
          ? '<ul id="sooauth-password-rules" style="list-style:none;padding:8px 0 0;margin:0;display:grid;gap:6px"></ul>'
          : "") +
        "</div>" +
         '<button id="sooauth-submit" type="submit" style="padding:12px;border:1px solid #051B23;border-radius:8px;background:' + accent + ';color:#051B23;font-weight:600;cursor:pointer">' +
        (mode === "signup" ? "Create account" : "Sign in") + "</button></form>" +
         '<p id="sooauth-msg" style="font-size:13px;color:#051B23;margin:8px 0 0"></p>';
    }
    if (cfg.show_powered_by) {
       html += '<p style="margin-top:16px;font-size:12px;color:#5D6B70;text-align:center">Powered by <a href="https://sooauth.com" style="color:#051B23">sooauth</a></p>';
    }
    html += "</div>";
    root.innerHTML = html;

    var form = document.getElementById("sooauth-email-form");
    if (!form) return;

    var passwordInput = document.getElementById("sooauth-password");
    var rulesEl = document.getElementById("sooauth-password-rules");
    var submitBtn = document.getElementById("sooauth-submit");
    var passwordTouched = false;

    function refreshPasswordUi() {
      if (!rulesEl || !passwordInput || mode !== "signup") return;
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

    if (passwordInput && mode === "signup") {
      passwordInput.addEventListener("input", refreshPasswordUi);
      passwordInput.addEventListener("focus", function () {
        passwordTouched = true;
        refreshPasswordUi();
      });
      refreshPasswordUi();
    }

    form.addEventListener("submit", function (e) {
      e.preventDefault();
      var fd = new FormData(form);
      var password = String(fd.get("password") || "");
      var msg = document.getElementById("sooauth-msg");
      msg.textContent = "";

      if (mode === "signup" && !meetsPolicy(password, policy)) {
        passwordTouched = true;
        refreshPasswordUi();
        return;
      }

      var path = mode === "signup" ? "/auth/sign-up" : "/auth/sign-in";
      fetch(base + path, {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-Sooauth-Embed": "1" },
        body: JSON.stringify({
          email: fd.get("email"),
          password: password,
          client_id: clientId,
        }),
      })
        .then(function (r) { return r.json().then(function (d) { return { ok: r.ok, d: d }; }); })
        .then(function (res) {
          if (!res.ok) {
             msg.style.color = "#051B23";
            msg.textContent = res.d.message || res.d.error || "Request failed";
            return;
          }
          if (mode === "signup") {
             msg.style.color = "#051B23";
            msg.textContent = res.d.message || "Check your email to verify your account.";
            return;
          }
          dispatch("sooauth:login", {
            access_token: res.d.access_token,
            email: res.d.email,
            expires_in: res.d.expires_in,
          });
           msg.style.color = "#051B23";
          msg.textContent = "Signed in.";
        });
    });
  }

  function boot(base) {
    loadConfig(base).then(function (cfg) {
      var params = new URLSearchParams(window.location.search);
      var widgetCode = params.get("widget_code");
      var err = params.get("error");
      if (err) {
         root.innerHTML = '<p style="font:14px system-ui;color:#051B23">Sign-in failed: ' + err + "</p>";
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
           root.innerHTML = '<p style="font:14px system-ui;color:#051B23">Signed in as ' + (res.d.email || "user") + ".</p>";
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
