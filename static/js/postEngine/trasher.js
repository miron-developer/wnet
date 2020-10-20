'use strict'


import { GetFocus, SetFocus } from "./cursor.js";
import { Photos, PhotosPlace } from "./generator.js";

export const RemoveElem = e => {
    if (e) e.stopPropagation();
    const index = GetFocus();
    if (index === 0) return;
    SetFocus(0);
    const elem = document.querySelector(`.post-body-element-${index}`);
    if (elem.classList.contains('post-image')) {
        PhotosPlace.delete(index)
        Photos.delete(index);
    }
    elem.remove();
}