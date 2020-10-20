'use strict'


import { Route, InitRoutes } from './spa/router.js';
import { IsUserLogged } from './logical/user.js';
import { CreateWSConnection } from './logical/ws.js';
import { Logout } from './logical/sign.js';


const init = async() => {
    InitRoutes();
    CreateWSConnection();
    await IsUserLogged();
}

document.addEventListener('DOMContentLoaded', async() => {
    await init();
    Route(window.location.pathname);
});

const menuBtn = document.querySelector('.menu-btn'),
    header = document.querySelector('.header'),
    main = document.querySelector('main'),
    mainLink = document.querySelector('.navs-main'),
    postsLink = document.querySelector('.navs-posts'),
    contactsLink = document.querySelector('.navs-contacts'),
    signLink = document.querySelector('.sign-btn'),
    logoLink = document.querySelector('.header-logo'),
    advancedSearchLink = document.querySelector('.navs-advanced-search'),
    search = document.querySelector('.header-search>input');


// TODO: search
search.addEventListener('keyup', e => {
    if (e.key === 'Enter')
        console.log(search.value, 'search'); // search method here
})

// control btn on mobile
menuBtn.addEventListener('click', () => {
    header.classList.add('open');
    if (header.classList.contains('close')) header.classList.remove('close');
});
main.addEventListener('click', () => {
    if (header.classList.contains('open')) {
        header.classList.remove('open');
        header.classList.add('close');
    }
});

// routes on main
mainLink.addEventListener('click', () => Route('/'));
logoLink.addEventListener('click', () => Route('/'));
postsLink.addEventListener('click', () => Route('/posts'));
contactsLink.addEventListener('click', () => Route('/contacts'));
signLink.addEventListener('click', () => (signLink.classList[1] === 'sign-in') ? Route('/sign/in') : Logout());
advancedSearchLink.addEventListener('click', () => Route('/advanced-search'));