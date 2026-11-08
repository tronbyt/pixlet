import { useEffect } from 'react';
import { useSelector, useDispatch } from 'react-redux';

import { callGeneratedHandler } from '../../handlers/actions';
import { set as setError } from '../../errors/errorSlice';

/**
 * Runs a generated declaration's handler using its configured source value.
 * Reruns when configuration or the base schema generation changes so a live
 * schema reload repopulates generated fields without a source-value edit.
 *
 * @param {Object} props - Component properties.
 * @param {Object} props.field - Generated declaration with ID, source, and handler.
 * @returns {null} This component triggers generation without rendering UI.
 */
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
