import OAuth2 from './fields/oauth2/OAuth2';
import PhotoSelect from './fields/photoselect/PhotoSelect';
import Toggle from './fields/Toggle';
import Color from './fields/Color';
import DateTime from './fields/DateTime';
import Dropdown from './fields/Dropdown';
import LocationBased from './fields/location/LocationBased';
import LocationForm from './fields/location/LocationForm';
import TextInput from './fields/TextInput';
import Typeahead from './fields/Typeahead';
import Typography from '@mui/material/Typography';


/**
 * Renders the input for one schema field in the pixlet serve config panel.
 *
 * @param {{ field: object }} props - the field, as serialized in the app's schema
 */
export default function FieldDetails({ field }) {
    switch (field.type) {
        case 'datetime':
            return <DateTime field={field} />
        case 'dropdown':
            return <Dropdown field={field} />
        case 'location':
            return <LocationForm field={field} />
        case 'locationbased':
            return <LocationBased field={field} />
        case 'oauth2':
            // Server-managed fields get their token from the Tronbyt server,
            // which pixlet serve is not; paste a token to test the app.
            if (field.token_endpoint) {
                return <TextInput field={{ ...field, name: `${field.name} access token` }} />
            }
            return <OAuth2 field={field} />
        case 'png':
            return <PhotoSelect field={field} />
        case 'text':
            return <TextInput field={field} />
        case 'onoff':
            return <Toggle field={field} />
        case 'typeahead':
            return <Typeahead field={field} />
        case 'color':
            return <Color field={field} />
        default:
            return <Typography>Unsupported type: {field.type}</Typography>
    }
}