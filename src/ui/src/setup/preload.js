import {
    SYSTEM_MANAGEMENT
} from '@/dictionary/auth'

import { viewRouters } from '@/router'

const preloadConfig = {
    fromCache: true,
    cancelWhenRouteChange: false
}

export function getSystemAuth (app) {
    return app.$store.dispatch('auth/getAuth', {
        type: 'system',
        list: [SYSTEM_MANAGEMENT],
        config: {
            ...preloadConfig,
            requestId: 'getSystemAuth'
        }
    })
}

export function getViewAuth (app) {
    const viewAuthorities = []
    viewRouters.forEach(route => {
        const meta = route.meta || {}
        const auth = meta.auth || {}
        if (auth.view) {
            viewAuthorities.push(auth.view)
        }
    })
    return app.$store.dispatch('auth/getAuth', {
        type: 'view',
        list: viewAuthorities,
        config: {
            ...preloadConfig,
            requestId: 'getViewAuth'
        }
    })
}

export function getClassifications (app) {
    return app.$store.dispatch('objectModelClassify/searchClassificationsObjects', {
        params: app.$injectMetadata(),
        config: {
            ...preloadConfig,
            requestId: 'post_searchClassificationsObjects'
        }
    })
}

export function getAuthorizedBusiness (app) {
    return app.$store.dispatch('objectBiz/getAuthorizedBusiness')
}

export function getUserCustom (app) {
    return app.$store.dispatch('userCustom/searchUsercustom', {
        config: {
            ...preloadConfig,
            fromCache: false,
            requestId: 'post_searchUsercustom'
        }
    })
}

export function getUserList (app) {
    return app.$store.dispatch('getUserList').then(list => {
        list = list || []
        window.CMDB_USER_LIST = list
        app.$store.commit('setUserList', list)
        return list
    }).catch(e => {
        window.CMDB_USER_LIST = []
    })
}

export default async function (app) {
    try {
        await Promise.all([
            getSystemAuth(app),
            getAuthorizedBusiness(app)
        ])
    } catch (e) {
        console.error(e)
    }
    return Promise.all([
        getViewAuth(app),
        getClassifications(app),
        getUserCustom(app),
        getUserList(app)
    ])
}
