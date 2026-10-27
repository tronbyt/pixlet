import { useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';

import { callGeneratedHandler } from '../../handlers/actions';
import { set as setError } from '../../errors/errorSlice';

export default function Generated({ field }) {
    const config = useSelector(state => state.config);
    const baseSchema = useSelector(state => state.schema.value);
    const generation = useSelector(state => state.schema.generation);
    const dispatch = useDispatch();

    const { id, source, handler } = field;

    useEffect(() => {
        const fields = baseSchema.schema;

        if (fields.length === 0) {
            return;
        }

        const sourceField = fields.find(
            candidate => candidate.id === source
        );

        if (!sourceField) {
            const msg =
                `schema.Generated references source that does not exist: ${source}`;

            dispatch(setError({ id: msg, message: msg }));
            return;
        }

        const sourceConfig = config[sourceField.id];

        if (
            !sourceConfig ||
            typeof sourceConfig.value === 'undefined'
        ) {
            return;
        }

        callGeneratedHandler(id, handler, sourceConfig.value);
    }, [config, baseSchema, generation, id, source, handler, dispatch]);

    return null;
}
