import { useEffect } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';

import fetchPreview from '../preview/actions';


export default function ConfigManager() {
    const config = useSelector(state => state.config);
    const loading = useSelector(state => state.param.loading);
    const preview = useSelector(state => state.preview);
    const renderScale = useSelector(state => state.preview.value.is2x);
    const metaTimezone = useSelector(state => state.preview.value.timezone);
    const metaLocale = useSelector(state => state.preview.value.locale);
    const navigate = useNavigate();

    const updatePreviews = (formData, params) => {
        navigate({ search: params.toString() });
        fetchPreview(formData);
    }

    useEffect(() => {
        const formData = new FormData();
        const params = new URLSearchParams();

        Object.entries(config).forEach((entry) => {
            const [id, item] = entry;

            // Not all config values fit inside a query parameter, most notably
            // images. If they don't fit, simply leave them out of the query
            // string. The downside is a refresh will lose that state.
            if (item.value.length < 1024) {
                params.set(id, item.value)
            }

            formData.set(id, item.value);
        });

        // metadata fields that should not collide with app config
        params.set('_metaTimezone', metaTimezone || '');
        formData.set('_metaTimezone', metaTimezone || '');
        params.set('_metaLocale', metaLocale || '');
        formData.set('_metaLocale', metaLocale || '');

        if (renderScale !== null && renderScale !== undefined) {
            const scaleValue = renderScale ? '2' : '1';
            params.set('_renderScale', scaleValue);
            formData.set('_renderScale', scaleValue);
        }

        if (!loading || !('img' in preview)) {
            updatePreviews(formData, params);
        }
    }, [config, renderScale, metaTimezone, metaLocale]);

    return null;
}
