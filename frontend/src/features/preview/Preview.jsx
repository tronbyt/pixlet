import { useSelector } from 'react-redux';

import Paper from '@mui/material/Paper';

import styles from './styles.module.css';
import loadingImage from './loading.webp';

export default function Preview() {
    const preview = useSelector(state => state.preview);

    let displayType = 'data:image/webp;base64,';
    if (preview.value.img_type === "gif") {
        displayType = 'data:image/gif;base64,';
    }

    let img = 'UklGRhoAAABXRUJQVlA4TA4AAAAvP8AHAAcQEf0PRET/Aw==';
    if (preview.value.img) {
        img = preview.value.img;
    }

    let dotsUrl = new URL('./api/v1/dots.svg', document.location);
    const scale = preview.value.is2x ? 2 : 1;
    if (preview.value.width) {
        dotsUrl.searchParams.set('w', String(preview.value.width * scale));
    }
    if (preview.value.height) {
        dotsUrl.searchParams.set('h', String(preview.value.height * scale));
    }

    return (
        <Paper sx={{ backgroundColor: "black", backgroundImage: 'none' }} className={styles.container}>
            {preview.loading ? (
                <img
                    src={loadingImage}
                    alt="Loading preview"
                    className={styles.image}
                    style={{ maskImage: `url("${dotsUrl}")`, WebkitMaskImage: `url("${dotsUrl}")` }}
                />
            ) : (
                <img
                    src={displayType + img}
                    className={styles.image}
                    style={{ maskImage: `url("${dotsUrl}")`, WebkitMaskImage: `url("${dotsUrl}")` }}
                />
            )}
        </Paper>
    );
}
