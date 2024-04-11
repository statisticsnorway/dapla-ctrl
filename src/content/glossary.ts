
interface AutonomyLevel {
    text: string
}

export const AUTONOMY_LEVEL: Record<string, AutonomyLevel> = {
    MANAGED: {
        text: "Team is Managed"
    },
    SELF_MANAGED: {
        text: "Self managed"
    },
    SEMI_MANAGED: {
        text: "Semi managed"
    },
    UNDEFINED: {
        text: "Autonomy level is undefined"
    }
}