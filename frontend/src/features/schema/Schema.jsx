import { Fragment, useEffect } from 'react';
import { useSelector } from 'react-redux';

import refreshSchema from './actions';
import Field from './Field';
import Generated from './fields/Generated';

export default function Schema() {
    const schema = useSelector(state => state.schema);

    useEffect(() => {
        refreshSchema();
    }, []);

    return (
        <div>
            {schema.value.schema.map((field) => {
                if (field.type === 'generated') {
                    const generatedFields =
                        schema.generated[field.id]?.schema ?? [];

                    return (
                        <Fragment key={field.id}>
                            <Generated field={field} />
                            {generatedFields.map((generatedField) => (
                                <Field
                                    key={generatedField.id}
                                    field={generatedField}
                                />
                            ))}
                        </Fragment>
                    );
                }

                return <Field key={field.id} field={field} />;
            })}
        </div>
    );
}
