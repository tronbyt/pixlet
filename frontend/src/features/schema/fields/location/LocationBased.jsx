import { useState, useEffect, useCallback } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import MenuItem from '@mui/material/MenuItem';
import FormControl from '@mui/material/FormControl';
import Select from '@mui/material/Select';
import Typography from '@mui/material/Typography';

import { set } from '../../../config/configSlice';
import { callHandler } from '../../../handlers/actions';

// Helper function to extract all URL parameters
const getUrlParams = () => {
    // Parse the query string from the URL
    const queryString = window.location.search;
    const params = new URLSearchParams(queryString);

    // Create an object with all parameters
    const urlParams = {};
    for (const [key, value] of params.entries()) {
        urlParams[key] = value;
    }

    return urlParams;
};

// UI component library for the location picker dropdown
// Base container for the entire dropdown component
const Command = ({ children, className, onClick }) => (
    <div className={`relative rounded-lg ${className}`} onClick={onClick}>
        {children}
    </div>
);

// Text input component for the search box
const CommandInput = ({ placeholder, value, onValueChange, onClick }) => (
    <input
        className="w-full px-4 py-2 rounded-lg border focus:outline-none focus:ring-2"
        type="text"
        placeholder={placeholder}
        value={value}
        onChange={(e) => onValueChange(e.target.value)}
        onClick={onClick}
    />
);

// Container for the results dropdown list
const CommandList = ({ children }) => (
    <div className="mt-1 max-h-60 overflow-auto rounded-lg border bg-white">
        {children}
    </div>
);

// Individual selectable item in the dropdown list
const CommandItem = ({ children, onSelect }) => (
    <div
        className="px-4 py-2 hover:bg-gray-100 cursor-pointer flex items-center"
        onClick={onSelect}
    >
        {children}
    </div>
);

// LocationPicker component that handles location search via Geoapify API
const LocationPicker = ({ onLocationSelect, initialLocation }) => {
    // State management for the search and results
    const [query, setQuery] = useState(initialLocation?.locality || '');
    const [results, setResults] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [isOpen, setIsOpen] = useState(false);

    // Debounce utility to prevent excessive API calls during typing
    const debounce = (func, wait) => {
        let timeout;
        return (...args) => {
            clearTimeout(timeout);
            timeout = setTimeout(() => func(...args), wait);
        };
    };

    // Function to search locations via Geoapify API
    const searchLocations = async (searchQuery) => {
        if (!searchQuery) {
            setResults([]);
            setIsOpen(false); // Close dropdown if query is empty
            return;
        }

        setLoading(true);
        setError(null);

        try {
            const apiKey = '49863bd631b84a169b347fafbf128ce6';
            const encodedQuery = encodeURIComponent(searchQuery);
            const response = await fetch(
                `https://api.geoapify.com/v1/geocode/search?text=${encodedQuery}&apiKey=${apiKey}`
            );

            if (!response.ok) {
                throw new Error('Failed to fetch locations');
            }

            const data = await response.json();

            // Transform the API response to our internal format
            const locations = data.features.map(feature => ({
                formattedName: feature.properties.formatted,
                locality: feature.properties.formatted,
                lat: feature.properties.lat,
                lng: feature.properties.lon,
                timezone: feature.properties.timezone?.name || 'America/New_York'
            }));

            setResults(locations);
        } catch (err) {
            setError('Failed to load locations');
            console.error('Error fetching locations:', err);
        } finally {
            setLoading(false);
        }
    };

    // Create a debounced search function to delay API calls
    const debouncedSearch = useCallback(
        debounce(searchLocations, 300),
        []
    );

    // Handler for search input changes
    const handleSearch = (value) => {
        setQuery(value);
        setIsOpen(true); // Open dropdown when searching
        debouncedSearch(value);
    };

    // Effect to handle clicks outside the component to close the dropdown
    useEffect(() => {
        const handleClickOutside = () => {
            setIsOpen(false);
        };

        // Add event listener when component mounts
        document.addEventListener('click', handleClickOutside);

        // Clean up the event listener on unmount
        return () => {
            document.removeEventListener('click', handleClickOutside);
        };
    }, []);

    // Prevent dropdown closure when clicking inside the component
    const handleCommandClick = (e) => {
        e.stopPropagation();
    };

    return (
        <Command className="w-full" onClick={handleCommandClick}>
            <CommandInput
                placeholder="Search for a location..."
                value={query}
                onValueChange={handleSearch}
                onClick={() => setIsOpen(true)} // Open dropdown when clicking input
            />
            {isOpen && loading && (
                <div className="p-4 text-center text-gray-500">
                    Loading...
                </div>
            )}
            {isOpen && error && (
                <div className="p-4 text-center text-red-500">
                    {error}
                </div>
            )}
            {isOpen && results.length > 0 && (
                <CommandList>
                    {results.map((location, index) => (
                        <CommandItem
                            key={index}
                            onSelect={() => {
                                onLocationSelect(location);
                                setIsOpen(false); // Close dropdown after selection
                                setResults([]); // Clear results
                            }}
                        >
                            <span className="mr-2">📍</span>
                            {location.formattedName}
                        </CommandItem>
                    ))}
                </CommandList>
            )}
        </Command>
    );
};

// Main component that extends the LocationForm with handler integration
export default function LocationBased({ field }) {
    // Get URL parameters to check for location data
    const urlParams = getUrlParams();

    // Try to find location data in the URL parameters
    let urlLocation = null;

    // First check if field.id exists as a parameter
    if (urlParams[field.id]) {
        try {
            const parsed = JSON.parse(decodeURIComponent(urlParams[field.id]));
            if (parsed && typeof parsed === 'object' && 'locality' in parsed) {
                urlLocation = parsed;
            }
        } catch (err) {
            console.error(`Error parsing ${field.id} from URL:`, err);
        }
    }

    // If location not found by field.id, search all URL parameters for valid location data
    if (!urlLocation) {
        for (const [key, value] of Object.entries(urlParams)) {
            try {
                const parsed = JSON.parse(decodeURIComponent(value));
                if (parsed && typeof parsed === 'object' && 'locality' in parsed) {
                    urlLocation = parsed;
                    break; // Found a valid location, stop searching
                }
            } catch (err) {
                // Not a valid JSON, continue to next parameter
                continue;
            }
        }
    }

    // Function to calculate default values with fallbacks and overrides
    const getDefaultValues = () => {
        const defaultValues = {
            // Default to Brooklyn as a fallback
            'lat': 40.678,
            'lng': -73.944,
            'locality': 'Brooklyn, New York',
            'timezone': 'America/New_York',
            'display': '', // For dropdown display value
            'value': '',   // For dropdown selected value
        };

        // Override with field.default if available
        if (field.default) {
            Object.assign(defaultValues, field.default);
        }

        // Override with URL location if available
        if (urlLocation) {
            // Merge the URL location with defaults for any missing fields
            return {
                ...defaultValues,
                ...urlLocation
            };
        }

        return defaultValues;
    };

    // Component state management
    const [value, setValue] = useState(getDefaultValues());
    const [initialized, setInitialized] = useState(false);

    // Redux hooks
    const config = useSelector(state => state.config);
    const dispatch = useDispatch();
    const handlerResults = useSelector(state => state.handlers);

    // Helper function to format location for handlers
    const getLocationAsJson = (v) => {
        return JSON.stringify({
            lat: v.lat,
            lng: v.lng,
            locality: v.locality,
            timezone: v.timezone
        });
    };

    // Effect to initialize the component
    useEffect(() => {
        // Set initialized to true
        setInitialized(true);
    }, []);

    // Effect to load configuration from Redux or initialize with defaults
    useEffect(() => {
        if (initialized && field.id in config) {
            // Load saved configuration from Redux
            let v = JSON.parse(config[field.id].value);
            setValue(v);

            // Call the associated handler with the location data
            callHandler(field.id, field.handler, getLocationAsJson(v));
        } else if (initialized) {
            // Use default values (including URL params if available)
            const defaultValues = getDefaultValues();

            // Save to Redux
            dispatch(set({
                id: field.id,
                value: JSON.stringify(defaultValues),
            }));
            setValue(defaultValues);

            // Call handler with initial location
            callHandler(field.id, field.handler, getLocationAsJson(defaultValues));
        }
    }, [config, dispatch, field.id, field.default, field.handler, initialized]);

    // Handler for location selection from the picker
    const handleLocationSelect = (location) => {
        // Create new state with selected location
        const newValue = {
            ...value,
            lat: location.lat.toString(),
            lng: location.lng.toString(),
            locality: location.locality || location.formattedName,
            timezone: location.timezone
        };

        // Update state and Redux store
        setValue(newValue);
        dispatch(set({
            id: field.id,
            value: JSON.stringify(newValue),
        }));

        // Call handler with new location to update location-dependent options
        callHandler(field.id, field.handler, getLocationAsJson(newValue));
    };

    // Handler for dropdown option selection
    const onChangeOption = (event) => {
        let newValue = { ...value };
        newValue.value = event.target.value;

        // Get display text from selected option
        const selectedOption = options.find(option => option.value === event.target.value);
        if (selectedOption) {
            newValue.display = selectedOption.display;
        }

        // Update state and Redux store
        setValue(newValue);
        dispatch(set({
            id: field.id,
            value: JSON.stringify(newValue),
        }));
    };

    // Get options from handler results for the dropdown
    let options = [];
    if (field.id in handlerResults.values) {
        options = handlerResults.values[field.id];
    }

    return (
        <FormControl fullWidth>
            {/* Location search field */}
            <Typography variant="subtitle1" gutterBottom>Location</Typography>
            <LocationPicker
                onLocationSelect={handleLocationSelect}
                initialLocation={value}
            />

            {/* Display selected location details */}
            {value && (
                <div className="mt-4 p-3 border rounded bg-gray-50">
                    <Typography variant="body2"><strong>Selected:</strong> {value.locality}</Typography>
                    <Typography variant="body2"><strong>Coordinates:</strong> {value.lat}, {value.lng}</Typography>
                    <Typography variant="body2"><strong>Timezone:</strong> {value.timezone}</Typography>
                </div>
            )}

            {/* Options dropdown for the chosen location */}
            <Typography style={{ marginTop: '1rem' }}>Options for chosen location</Typography>
            <Select
                onChange={onChangeOption}
                value={value['value']}
                fullWidth
            >
                {options.map((option) => (
                    <MenuItem key={option.value} value={option.value}>{option.display}</MenuItem>
                ))}
            </Select>
        </FormControl>
    );
}