package web

import (
	"encoding/json"
	"fmt"

	"ticketd/internal/store"
)

// buildEmbedJS generates the JavaScript code for embedding a form on external websites.
// The generated script is a self-contained IIFE that creates a form widget with:
// - CSS loading (from the configured base URL)
// - Form field generation based on form type (contact/support)
// - CORS-enabled form submission handling
// - Success/error status display
//
// The script can be embedded using a <script> tag: <script src="https://yourserver.com/embed/{formID}.js"></script>
func buildEmbedJS(form store.Form, client store.Client, baseURL string) (string, error) {
	cssURL := fmt.Sprintf("%s/embed/form.css", baseURL)
	apiURL := fmt.Sprintf("%s/api/forms/%d/submit", baseURL, form.ID)
	formTitle := fmt.Sprintf("%s - %s", client.Name, form.Name)

	// Build form fields based on form type
	fields := []map[string]any{
		{"label": "Name", "name": "name", "type": "text"},
		{"label": "Email", "name": "email", "type": "email"},
	}
	if form.Type == store.FormTypeSupport {
		fields = append(fields, map[string]any{"label": "Subject", "name": "subject", "type": "text"})
		fields = append(fields, map[string]any{
			"label":   "Priority",
			"name":    "priority",
			"type":    "select",
			"options": []string{"low", "medium", "high"},
		})
	}
	fields = append(fields, map[string]any{"label": "Message", "name": "message", "type": "textarea"})

	payload := map[string]any{
		"cssURL":   cssURL,
		"apiURL":   apiURL,
		"title":    formTitle,
		"fields":   fields,
		"formType": string(form.Type),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Generate the self-contained JavaScript embed code
	script := fmt.Sprintf(`(function(){
  var cfg = %s;
  var scriptTag = document.currentScript;
  var mount = document.createElement("div");
  mount.className = "ticketd-embed";
  if (scriptTag && scriptTag.parentNode) {
    scriptTag.parentNode.insertBefore(mount, scriptTag);
  } else {
    document.body.appendChild(mount);
  }
  if (!document.querySelector('link[data-ticketd="true"]')) {
    var link = document.createElement("link");
    link.rel = "stylesheet";
    link.href = cfg.cssURL;
    link.setAttribute("data-ticketd", "true");
    document.head.appendChild(link);
  }

  var form = document.createElement("form");
  form.className = "ticketd-form";
  var title = document.createElement("h3");
  title.textContent = cfg.title;
  form.appendChild(title);

  cfg.fields.forEach(function(field){
    var label = document.createElement("label");
    label.textContent = field.label;
    var input;
    if (field.type === "textarea") {
      input = document.createElement("textarea");
      input.rows = 4;
    } else if (field.type === "select") {
      input = document.createElement("select");
      field.options.forEach(function(opt){
        var option = document.createElement("option");
        option.value = opt;
        option.textContent = opt;
        input.appendChild(option);
      });
    } else {
      input = document.createElement("input");
      input.type = field.type || "text";
    }
    input.name = field.name;
    input.required = true;
    form.appendChild(label);
    form.appendChild(input);
  });

  var button = document.createElement("button");
  button.type = "submit";
  button.textContent = "Send";
  form.appendChild(button);

  var status = document.createElement("div");
  status.className = "ticketd-status";
  form.appendChild(status);

  form.addEventListener("submit", function(event){
    event.preventDefault();
    status.textContent = "Sending...";
    status.className = "ticketd-status";
    var payload = {};
    Array.prototype.forEach.call(form.elements, function(el){
      if (!el.name || el.type === "submit") {
        return;
      }
      payload[el.name] = el.value;
    });
    fetch(cfg.apiURL, {
      method: "POST",
      mode: "cors",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload)
    })
      .then(function(res){ return res.json().then(function(body){ return { ok: res.ok, body: body }; }); })
      .then(function(result){
        if (!result.ok) {
          throw new Error(result.body && result.body.error ? result.body.error : "Failed");
        }
        status.textContent = "Thanks! We'll be in touch.";
        status.className = "ticketd-status ticketd-success";
        form.reset();
      })
      .catch(function(err){
        status.textContent = err.message || "Failed to send.";
        status.className = "ticketd-status ticketd-error";
      });
  });

  mount.appendChild(form);
})();`, string(data))

	return script, nil
}
