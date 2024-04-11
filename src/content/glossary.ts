
interface AutonomyLevel {
    text: string
}

export const AUTONOMY_LEVEL: Record<string, AutonomyLevel> = {
    MANAGED: {
        text: "Et managed Dapla-team er et team som kun kan bruke tjenester som tilbys på Dapla."
    },
    SEMI_MANAGED: {
        text: "Et semi-managed Dapla-team er et team som stort sett benytter seg tjenestene som tilbys på Dapla, men de har også noe frihet til å ta ansvar for deler av sin egen infrastruktur."
    },
    SELF_MANAGED: {
        text: "Et self-managed Dapla-team er et team som står helt fritt til å definere sin egen infrastruktur og er ansvarlig for at den er satt opp iht SSBs krav."
    },
    UNDEFINED: {
        text: "Autonomy level is undefined"
    }
}