(function(window) {
  "use strict";

  const Name = props => {
    return input({
      name: props.name ? props.name : "",
      className: "form-control mr-2 col-md-5",
      type: "text",
      value: props.value ? props.value : "",
      placeholder: "NAME",
      readOnly: props.readOnly,
      onchange: props.onchange ? props.onchange : null
    });
  };

  const Value = props => {
    return textarea({
      name: props.name ? props.name : "",
      className: "form-control mr-2 col-md-5",
      value: props.value ? props.value : "",
      placeholder: "VALUE",
      readOnly: props.readOnly,
      wrap: "off",
      rows: 1,
      style: { resize: "none", overflow: "hidden" },
      onchange: props.onchange ? props.onchange : null
    });
  };

  const DeleteButton = props => {
    return svg(
      {
        style: { width: "22px", height: "22px" },
        onclick: props.onclick
      },
      path({
        d:
          "M6.66322227,8.01000023 L3.30088704,11.3723355 C2.92587893,11.7473436 2.92437491,12.3518206 3.29627738,12.7237231 C3.67077268,13.0982184 4.27321669,13.0935617 4.647665,12.7191134 L8.01000023,9.35677819 L11.3723355,12.7191134 C11.7473436,13.0941215 12.3518206,13.0956255 12.7237231,12.7237231 C13.0982184,12.3492278 13.0935617,11.7467838 12.7191134,11.3723355 L9.35677819,8.01000023 L12.7191134,4.647665 C13.0941215,4.27265689 13.0956255,3.66817984 12.7237231,3.29627738 C12.3492278,2.92178208 11.7467838,2.92643873 11.3723355,3.30088704 L8.01000023,6.66322227 L4.647665,3.30088704 C4.27265689,2.92587893 3.66817984,2.92437491 3.29627738,3.29627738 C2.92178208,3.67077268 2.92643873,4.27321669 3.30088704,4.647665 L6.66322227,8.01000023 L6.66322227,8.01000023 Z",
        style: { fill: "red", stroke: "#cc0000" }
      })
    );
  };

  const ConfigVars = props => {
    const state = {
      name: "",
      value: ""
    };

    const handleClick = e => {
      e.preventDefault();
      props.onAdd([state.name, state.value]);
      const div = document.querySelector("div.config-vars div:last-child");
      const input = div.querySelector("input");
      input.value = "";
      input.focus();
      div.querySelector("textarea").value = "";
    };

    const handleChange = e => {
      const target = e.target;
      const btn = target.parentElement.querySelector("button");
      if (target.tagName === "INPUT") {
        state.name = target.value.trim();
      } else if (target.tagName === "TEXTAREA") {
        state.value = target.value.trim();
      }
      btn.disabled = state.name === "";
    };

    const handleDelete = i => {
      props.onDelete(i);
    };

    const list = props.pairs.map((pair, i) => {
      return div(
        { className: "config-var-pair form-inline mb-2" },
        Name({
          name: "configuration[names][]",
          value: pair[0],
          readOnly: true
        }),
        Value({
          name: "configuration[values][]",
          value: pair[1],
          readOnly: true
        }),
        DeleteButton({ onclick: () => handleDelete(i) })
      );
    });

    return div(
      { className: "form-group required" },
      label({ className: "col-form-label" }, "Config vars"),
      div(
        { className: "config-vars" },
        list,
        div(
          { className: "config-var-pair form-inline" },
          Name({ onchange: e => handleChange(e) }),
          Value({ onchange: e => handleChange(e) }),
          button(
            {
              className: "btn btn-secondary",
              onclick: e => handleClick(e),
              disabled: true
            },
            "Add"
          )
        )
      )
    );
  };

  function empty(el) {
    while (el.firstChild) {
      el.removeChild(el.firstChild);
    }
  }

  function ConfigVarsApp(el, initial) {
    initial = initial || [];
    let pairs = [];
    let props = {
      pairs: pairs,
      onAdd: pair => {
        let found = false;
        for (let i = 0; i < pairs.length; i++) {
          if (pairs[i][0] === pair[0]) {
            pairs[i][1] = pair[1];
            found = true;
            break;
          }
        }
        if (!found) {
          pairs.push(pair);
        }
        render();
      },
      onDelete: i => {
        pairs.splice(i, 1);
        render();
      }
    };
    initial.forEach(pair => {
      props.pairs.push(pair);
    });
    const render = () => {
      empty(el);
      el.appendChild(ConfigVars(props));
      el.appendChild(
        SaveBtn({
          disabled: () => !pairs.length || arraysEqual(initial, pairs)
        })
      );
      el.appendChild(
        CancelBtn({
          disabled: () => arraysEqual(initial, pairs),
          onCancel: () => {
            pairs = props.pairs = initial;
            render();
          }
        })
      );
    };
    render();
  }

  const SaveBtn = props => {
    return button(
      {
        className: "btn btn-primary",
        disabled: props.disabled()
      },
      "Save"
    );
  };

  const CancelBtn = props => {
    return button(
      {
        className: "btn btn-secondary",
        onclick: e => {
          e.preventDefault();
          props.onCancel();
        },
        disabled: props.disabled()
      },
      "Cancel"
    );
  };

  window.ConfigVarsApp = ConfigVarsApp;

  function arraysEqual(a, b) {
    if (!a || !b) return false;
    if (a.length !== b.length) return false;
    for (let i = 0; i < a.length; i++) {
      if (a[i] instanceof Array && b[i] instanceof Array) {
        if (!arraysEqual(a[i], b[i])) {
          return false;
        }
      } else if (a[i] !== b[i]) {
        return false;
      }
    }
    return true;
  }
})(window);
