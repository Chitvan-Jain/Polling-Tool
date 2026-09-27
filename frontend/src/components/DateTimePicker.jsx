import { useEffect, useRef } from "react";
import flatpickr from "flatpickr";
import "flatpickr/dist/flatpickr.min.css";

export function DateTimePicker({ value, onChange, minDate }) {
  const inputRef = useRef(null);
  const fpRef = useRef(null);

  useEffect(() => {
    fpRef.current = flatpickr(inputRef.current, {
      enableTime: true,
      dateFormat: "M j, Y — h:i K",
      minDate,
      defaultDate: value || undefined,
      onChange: (selectedDates) => {
        onChange(selectedDates[0] || null);
      },
    });

    return () => {
      fpRef.current?.destroy();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return <input ref={inputRef} type="text" readOnly placeholder="Choose a date and time" />;
}