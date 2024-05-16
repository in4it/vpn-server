import { Alert } from "@mantine/core";
import { IconInfoCircle } from "@tabler/icons-react";
import { useSearchParams } from "react-router-dom";

export function AuthError() {
    const alertIcon = <IconInfoCircle />;

    let [ searchParams, _ ] = useSearchParams();
    if(!searchParams.has("error")) return ''
    return (
        <Alert style={{marginTop: 20}} variant="light" color="red" title="Error" icon={alertIcon}>{searchParams.has("error_description") ? "An error occured: "+searchParams.get("error") +": " + searchParams.get("error_description"): "An error occured: "+searchParams.get("error")}</Alert>
    )
}