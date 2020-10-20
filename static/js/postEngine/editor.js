'use strict'


import { GetFocus } from "./cursor.js";

export const SetAlign = e => {
    e.stopPropagation();
    let align = e.path[1].classList[0].split('-')[1];
    if (align === "aligns") align = e.path[0].classList[0].split('-')[1];
    const index = GetFocus();
    if (index === 0) return;
    const elem = document.querySelector(`.post-body-element-${index}`);
    const arr = Array.from(elem.classList);
    elem.classList.remove(arr.find(cls => /align-\w+/g.test(cls)));
    elem.classList.add(`align-${align}`);
}

export const SetDecor = e => {
    e.stopPropagation();
    let decor = e.path[1].classList[0].split('-')[1];
    if (decor === "decors") decor = e.path[0].classList[0].split('-')[1];
    const index = GetFocus();
    if (index === 0) return;
    const elem = document.querySelector(`.post-body-element-${index}`);
    elem.classList.toggle('text-' + decor);
}

export const SetFS = e => {
    e.stopPropagation();
    const index = GetFocus();
    if (index === 0) return;
    const elem = document.querySelector(`.post-body-element-${index}`);
    const arr = Array.from(elem.classList)
    elem.classList.remove(arr.find(cls => /font-size-\w+/g.test(cls)));
    elem.classList.add(`font-size-${e.target.value}`);
}