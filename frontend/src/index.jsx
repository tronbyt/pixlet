import { lazy } from 'react';
import { Provider } from 'react-redux';
import { BrowserRouter, Route, Routes } from 'react-router-dom';
import { createRoot } from 'react-dom/client';

import store from './store';

const DevToolsTheme = lazy(() => import('./features/theme/DevToolsTheme'));
const Main = lazy(() => import('./Main'));
const OAuth2Handler = lazy(() => import('./features/schema/fields/oauth2/OAuth2Handler'));

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
