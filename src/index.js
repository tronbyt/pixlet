import { lazy } from 'react';
import { Provider } from 'react-redux';
import { BrowserRouter, Route, Routes } from 'react-router-dom';
import { createRoot } from 'react-dom/client';

import store from './store';

const DevToolsTheme = lazy(() => import(/* webpackChunkName: "devtoolstheme" */ './features/theme/DevToolsTheme'));
const Main = lazy(() => import(/* webpackChunkName: "main" */ './Main'));
const OAuth2Handler = lazy(() => import(/* webpackChunkName: "oauth2handler" */ './features/schema/fields/oauth2/OAuth2Handler'));

const App = () => {
    return (
        <Provider store={store}>
            <DevToolsTheme>
                <BrowserRouter basename={window.location.pathname}>
                    <Routes>
                        <Route exact path="/" element={<Main />} />
                        <Route path="oauth-callback" element={<OAuth2Handler />} />
                    </Routes>
                </BrowserRouter>
            </DevToolsTheme>
        </Provider>
    )
}

const container = document.getElementById('app');
const root = createRoot(container);
root.render(<App />);