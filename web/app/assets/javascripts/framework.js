(function(window) {
  "use strict";

  function appendText(el, text) {
    const textNode = document.createTextNode(text);
    el.appendChild(textNode);
  }

  function appendArray(el, children) {
    children.forEach(child => {
      if (Array.isArray(child)) {
        appendArray(el, child);
      } else if (child instanceof window.Element) {
        el.appendChild(child);
      } else if (typeof child === "string") {
        appendText(el, child);
      }
    });
  }

  function setStyles(el, styles) {
    Object.keys(styles).forEach(name => {
      if (name in el.style) {
        el.style[name] = styles[name];
      } else {
        console.warn(`invalid style for ${el.tagName}: ${name}`);
      }
    });
  }

  const svgTags = ["path", "rect", "svg"];

  function makeElement(type, props, ...rest) {
    let el;
    if (svgTags.indexOf(type) >= 0) {
      el = document.createElementNS("http://www.w3.org/2000/svg", type);
    } else {
      el = document.createElement(type);
    }

    if (Array.isArray(props)) {
      appendArray(el, props);
    } else if (props instanceof window.Element) {
      el.appendChild(props);
    } else if (typeof props === "string") {
      appendText(el, props);
    } else if (typeof props === "object") {
      Object.keys(props).forEach(name => {
        const value = props[name];
        if (name in el || name === "role") {
          if (name === "style") {
            setStyles(el, value);
          } else {
            el[name] = props[name];
          }
        } else {
          if (svgTags.indexOf(type) >= 0) {
            el.setAttribute(name, value);
          } else {
            console.warn(`invalid property of ${type}: ${name}`);
          }
        }
      });
    }

    if (rest) {
      appendArray(el, rest);
    }

    return el;
  }

  const tagNames = [
    "a",
    "b",
    "button",
    "div",
    "h1",
    "header",
    "input",
    "label",
    "p",
    "span",
    "textarea"
  ].concat(svgTags);
  tagNames.forEach(name => {
    const tag = (...args) => makeElement(name, ...args);
    window[name] = tag;
  });
})(window);
