'use strict'


import { SetFocus } from "./cursor.js";
import { InformUser } from '../spa/inform.js';

let pbody;
let index = 0;
export const Photos = new Map();
export const PhotosPlace = new Map();

export const InitPBody = () => {
    pbody = document.querySelector('.create-post-body');
    Photos.clear();
    index = 0;
    document.body.addEventListener('click', e => {
        e.stopPropagation();
        SetFocus(0);
    });
}

export const CreateList = list => {
    pbody.insertAdjacentHTML('beforeend', `<${list} class="post-body-element-${++index} post-list-${list}"></${list}>`);
    SetFocus(index);
}

export const CreateListNode = (elem, value) => {
    elem.insertAdjacentHTML('beforeend', `<li class="post-list-item post-body-element-${++index}">${value}</li>`);
    return index;
}

const isHave = file => {
    for (let f of Photos)
        if (f[1].name === file.name) return true;
    return false;
}

export const CreateImage = (imageSrc, file) => {
    pbody.insertAdjacentHTML('beforeend', ` <div class="post-body-element-${++index} post-image">
                                                <img src=${imageSrc}></img>
                                            </div>`);
    if (!isHave(file)) {
        if (Photos.size > 5) return InformUser('Max exceeded');
        Photos.set(index, file);
    }
    PhotosPlace.set(index, file.name);
    return index;
}

export const CreateSpoiler = e => {
    e.stopPropagation();
    pbody.insertAdjacentHTML('beforeend', ` <div class="post-body-element-${++index} post-spoiler"></div>`);
    SetFocus(index);
}

export const CreateText = e => {
    e.stopPropagation();
    pbody.insertAdjacentHTML('beforeend', `<div class="post-body-element-${++index}"></div>`);
    SetFocus(index);
}

export const CreateCitate = e => {
    e.stopPropagation();
    pbody.insertAdjacentHTML('beforeend', `<code class="post-body-element-${++index} post-citate"></code>`);
    SetFocus(index);
}