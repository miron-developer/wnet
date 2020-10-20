'use strict'


import { GeneratePreloader } from './content.js';
import { MainPage, ContactsPage, PostsPage, SignInPage, SignUpPage, RestorePage, RestorePasswordPage, SaveUserPage, AdvancedSearchPage, NotFoundPage, PostPage, ChangeProfilePage, CreatePostPage, MyPostsPage, ChatPage, UserProfilePage } from './pages.js';

let prevPage = '';
const routes = new Map();

export const AddRoute = (path, func) => routes.set(path, func);
export const AddRoutes = (paths = [], funcs = []) =>
    paths.forEach((v, i) => AddRoute(v, funcs[i]));

// TODO: advanced search
export const InitRoutes = () => {
    AddRoutes(
        ['/', '/contacts', '/posts', '/sign/in', '/sign/up', '/sign/restore',
            /\/post\/\w+/, '/404', /\/sign\/s\/\w*/, /\/sign\/r\/\w*/, '/advanced-search',
            '/profile/change-profile', '/create-post', '/profile/my-posts', /\/chat\/\w*/, /\/user\/\w*/,
        ], [MainPage, ContactsPage, PostsPage, SignInPage, SignUpPage, RestorePage,
            PostPage, NotFoundPage, SaveUserPage, RestorePasswordPage, AdvancedSearchPage,
            ChangeProfilePage, CreatePostPage, MyPostsPage, ChatPage, UserProfilePage
        ]
    );
}

// handle back and forward btn
window.onpopstate = e => Route(e.state);

// routes
export const Route = URL => {
    if (URL === prevPage) return;
    prevPage = URL;
    history.pushState(URL, '', URL);
    GeneratePreloader(document.querySelector('.content'));

    for (let [path, func] of routes) {
        if (path instanceof RegExp) {
            if (path.test(URL)) return func();
            continue;
        }
        if (path === URL) return func();
    }
    return routes.get('/404')();
}