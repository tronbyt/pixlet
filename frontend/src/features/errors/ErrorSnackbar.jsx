import { createRef } from 'react';
import { SnackbarProvider } from 'notistack';
import styles from './styles.module.css';

export default function ErrorSnackbar(props) {
    const notistackRef = createRef();

    return (
        <SnackbarProvider
            hideIconVariant
            preventDuplicate
            maxSnack={10}
            ref={notistackRef}
            classes={{
                variantError: styles.error,
            }}
            anchorOrigin={{
                vertical: 'bottom',
                horizontal: 'center',
            }}
        >
            {props.children}
        </SnackbarProvider>
    )
};
