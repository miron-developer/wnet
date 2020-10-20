'use strict'


import { InformUser } from "../spa/inform.js";
import { CreateImage, CreateList, CreateListNode } from "./generator.js";
import { RemoveElem } from "./trasher.js";

let currentElem = 0;
export const SetFocus = index => {
    if (currentElem === index) return;
    if (currentElem !== 0) SaveElem(currentElem);
    if (index !== 0) EditElem(index);
    currentElem = index;
}

export const GetFocus = () => currentElem;

const getClasses = elem => Array.from(elem.classList).join(' ');

// thx bro for this funcs
const addAutoResize = index => {
    if (!index || index <= 0) return;

    const element = document.querySelector(`.post-body-element-${index}`);
    const offset = element.offsetHeight - element.clientHeight;
    element.style.boxSizing = 'border-box';
    element.style.height = element.scrollHeight + offset + 'px';

    element.addEventListener('input', e => e.target.style.height = e.target.scrollHeight + offset + 'px');
}

const handleList = elem => {
    elem.insertAdjacentHTML('afterend', `<input class="post-body-li"></input>`);
    const input = document.querySelector('.post-body-li');
    input.focus();
    input.addEventListener('keydown', e => {
        if (e.key === 'Enter' && input.value !== '') {
            const index = CreateListNode(elem, input.value);
            document.querySelector(`.post-body-element-${index}`).addEventListener('click', e => {
                e.stopPropagation();
                SetFocus(index);
            });
            input.value = '';
        }
    });
}

export const HandleImage = () => {
    const input = document.createElement('input');
    input.setAttribute('type', 'file');
    input.click();

    input.addEventListener('change', async(e) => {
        e.stopPropagation();
        const file = input.files[0];
        if (file) {
            const reader = new FileReader();
            reader.readAsDataURL(file);
            reader.addEventListener('load', e => {
                const index = CreateImage(e.target.result, file);
                document.querySelector(`.post-body-element-${index}`).addEventListener('click', e => {
                    e.stopPropagation();
                    SetFocus(index);
                });
            });
        }
    });
}

export const HandleList = e => {
    e.stopPropagation();
    let list = e.path[1].classList[0].split('-')[1];
    if (list === "lists") list = e.path[0].classList[0].split('-')[1];
    CreateList(list);
}

export const EditElem = index => {
    const elem = document.querySelector(`.post-body-element-${index}`);
    if (elem.classList.contains('post-image')) return;
    if (elem.classList.contains('post-list-ul') || elem.classList.contains('post-list-ol'))
        return handleList(elem);

    elem.insertAdjacentHTML('beforebegin', `<textarea class="${getClasses(elem)}">${elem.textContent}</textarea>`)
    elem.remove();
    const newElem = document.querySelector(`.post-body-element-${index}`);
    newElem.addEventListener('click', e => {
        e.stopPropagation();
        SetFocus(index);
    });
    addAutoResize(index);
    newElem.setSelectionRange(newElem.value.length, newElem.value.length);
    newElem.focus();
}

export const CheckElemContent = content => /<+/g.test(content) || /\/+/g.test(content) || />+/g.test(content);

export const SaveElem = index => {
    const elem = document.querySelector(`.post-body-element-${index}`);
    if (elem.className === 'post-image' || elem.classList.contains('post-image')) return;
    if (elem.classList.contains('post-list-ul') || elem.classList.contains('post-list-ol'))
        return document.querySelector('.post-body-li').remove();
    if (elem.classList.contains('post-spoiler'))
        elem.insertAdjacentHTML('beforebegin', `<div class="${getClasses(elem)}">${elem.value}</div>`);
    else if (elem.classList.contains('post-citate'))
        elem.insertAdjacentHTML('beforebegin', `<code class="${getClasses(elem)}">${elem.value}</code>`);
    else if (elem.classList.contains('post-list-item'))
        elem.insertAdjacentHTML('beforebegin', `<li class="${getClasses(elem)}">${elem.value}</li>`);
    else
        elem.insertAdjacentHTML('beforebegin', `<div class="${getClasses(elem)}">${elem.value}</div>`);

    elem.remove();
    if (CheckElemContent(elem.value)) {
        InformUser("invalid text");
        return RemoveElem();
    }

    document.querySelector(`.post-body-element-${index}`).addEventListener('click', e => {
        e.stopPropagation();
        SetFocus(index);
    });
}