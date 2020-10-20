'use strict'


import { HandleImage, HandleList, SetFocus } from "./cursor.js";
import { SetAlign, SetDecor, SetFS } from "./editor.js";
import { CreateCitate, CreateSpoiler, CreateText, InitPBody, Photos, PhotosPlace } from "./generator.js";
import { RemoveElem } from './trasher.js';
import { Fetching } from '../spa/fetch.js';
import { InformUser } from '../spa/inform.js';
import { AddPost } from "../spa/pages.js";
import { Route } from "../spa/router.js";

const removeEmptyElements = () => {
    const all = document.querySelectorAll('.create-post-body > *');
    SetFocus(0);
    all.forEach(elem => elem.childNodes.length === 0 ? elem.remove() : null);
}

const saveImage = file => async() => {
    const data = new FormData();
    data.append('img', file);
    const res = await Fetching('/save/image', 'POST', data);
    if (res.msg !== 'ok') return InformUser('wrong save image');

    PhotosPlace.forEach((ph, index) => {
        if (ph === file.name)
            document.querySelector(`.post-body-element-${index} > img`).setAttribute('src', '/' + res.fname);
    });
}

const saveImages = async() => {
    const savingArr = [];
    Photos.forEach(file => savingArr.push(saveImage(file)));
    for (let v of savingArr) await v()
}

const handleTags = (alltags = []) => {
    const res = [];
    alltags.forEach(tag => (tag.value !== '' && !res.includes(tag.value)) ? res.push(tag.value) : null);
    return res
}

const savePost = async(e) => {
    e.stopPropagation();
    const title = document.querySelector('.create-post-title > input').value;
    if (!title) return InformUser('empty title');
    const tags = handleTags(
        Array.from(document.querySelectorAll('.post-body-wrapper input[name="tag"]'))
    )
    if (tags.length === 0) return InformUser('empty tag');
    const body = document.querySelector('.create-post-body');
    if (body.innerHTML === '') return InformUser('empty body');

    removeEmptyElements();
    await saveImages();

    const data = new FormData();
    data.append('tags', tags.join(','));
    data.append('title', title);
    data.append('body', body.innerHTML);
    const res = await Fetching('/save/post', 'POST', data);
    if (res.msg !== 'ok') InformUser(res.msg);
    else {
        AddPost('', res.post);
        Route('/');
    }
}

export const InitPE = () => {
    document.querySelector('.tools-alignment').addEventListener('click', e => e.stopPropagation());
    document.querySelector('.tools-text-decoration').addEventListener('click', e => e.stopPropagation());
    document.querySelector('.font-sizing').addEventListener('click', e => e.stopPropagation());

    document.querySelector('.align-left').addEventListener('click', SetAlign);
    document.querySelector('.align-right').addEventListener('click', SetAlign);
    document.querySelector('.align-center').addEventListener('click', SetAlign);
    document.querySelector('.align-justify').addEventListener('click', SetAlign);

    document.querySelector('.decors-bold').addEventListener('click', SetDecor);
    document.querySelector('.decors-italic').addEventListener('click', SetDecor);
    document.querySelector('.decors-underline').addEventListener('click', SetDecor);

    document.querySelector('.sizing-font-size').addEventListener('input', SetFS);

    document.querySelector('.list-ol').addEventListener('click', HandleList);
    document.querySelector('.list-ul').addEventListener('click', HandleList);
    document.querySelector('.elem-image').addEventListener('click', HandleImage);
    document.querySelector('.elem-spoiler').addEventListener('click', CreateSpoiler);
    document.querySelector('.elem-text').addEventListener('click', CreateText);
    document.querySelector('.elem-citate').addEventListener('click', CreateCitate);

    document.querySelector('.save').addEventListener('click', savePost);
    document.querySelector('.delete-current-elem').addEventListener('click', RemoveElem);
    InitPBody();
}