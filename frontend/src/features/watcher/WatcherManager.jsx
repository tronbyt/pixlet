import { useEffect } from 'react';

import Watcher from './watcher';


export default function WatcherManager() {
    useEffect(() => {
        const urlParams = new URLSearchParams(window.location.search);
        if (urlParams.get('$watch') === 'false') {
            return;
        }
        new Watcher();
    }, []);

    return null;
}