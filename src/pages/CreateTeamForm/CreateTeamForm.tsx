import styles from './createTeamForm.module.scss'

import {
    Button,
    Card,
    CheckboxGroup,
    Dialog,
    Dropdown,
    Glossary,
    Input,
    Text,
    TextArea,
} from '@statisticsnorway/ssb-component-library'
import * as C from '@statisticsnorway/ssb-component-library'
import { Skeleton } from '@mui/material'
import { useEffect, useState, useMemo } from 'react'
import { Array as A, Console, Effect, Option as O, pipe } from 'effect'

import FormSubmissionResult, { FormSubmissionResultProps } from './FormSubmissionResult.tsx'
import PageLayout from '../../components/PageLayout/PageLayout'
import * as Klass from '../../services/klass'
import { AutonomyLevel, CreateTeamRequest, createTeam } from '../../services/createTeam'
import { User } from '../../@types/user'
import * as Utils from '../../utils/utils.ts'

interface DisplayAutonomyLevel {
    id: AutonomyLevel
    title: string
}

interface DisplaySSBSection {
    id: number
    title: string
}

interface FormError {
    id: number
    field: string
    errorMessage: string
}

const CreateTeamForm = () => {
    const uniformNameLengthLimit = 17
    // TODO: These should be fetched from the dapla-team-api instead of being hardcoded
    const teamAutonomyLevels: DisplayAutonomyLevel[] = [
        { id: 'managed', title: 'Managed' },
        { id: 'semi-managed', title: 'Semi-Managed' },
        { id: 'autonomous', title: 'Self-Managed' },
    ]
    const teamNameGlossaryExplanation = `
      Teamets navn i et lesevennlig format. Navnet bør bestå av et hoveddomenet og et subdomenet, f.eks. "Skatt Næring" og det er tillatt med mellomrom og norske tegn (Æ, Ø, Å).
    `
    const uniformNameGlossaryExplanation = `
      Teamets navn i et maskinvennlig format som bl.a. benyttes i filstier til lagringsbøtter. Det er ikke tillatt med mellomrom og norske tegn (Æ, Ø, Å).
    `

    const sectionGlossaryExplanation = 'Ansvarlig seksjon for teamet.'

    const autonomyLevelGlossaryExplanation = `
      Nivå av frihet et team har til å definere sin egen infrastruktur. Statistikkproduserende team er vanligvis i kategorien "Managed", mens IT-team er "Self Managed". Les mer her.
    `

    const additionalInformationGlossaryExplanation = `
      Informasjon som kan være nyttig for den som oppretter teamet. F.eks. kan man liste opp hvem som skal legges i tilgangsgruppene data-admins og developers her.
`

    const displayNameLabel = 'Visningsnavn'
    const [displayName, setDisplayName] = useState('')

    const uniformNameLabel = 'Teknisk teamnavn'
    const [uniformName, setUniformName] = useState<string>('')
    const [overrideUniformName, setOverrideUniformName] = useState(false)
    const [uniformNameErrorMsg, setUniformNameErrorMsg] = useState<string>('')

    const sectionLabel = 'Eierseksjon'
    const [sections, setSections] = useState<DisplaySSBSection[]>([])
    const [selectedSection, setSelectedSection] = useState<O.Option<DisplaySSBSection>>(O.none())

    const [user, setUser] = useState<O.Option<User>>(O.none)

    const [selectedAutonomyLevel, setSelectedAutonomyLevel] = useState<DisplayAutonomyLevel>(teamAutonomyLevels[0])

    const [additionalInformation, setAdditionalInformation] = useState('')

    const [submitButtonClicked, setSubmitButtonClicked] = useState(false)

    const missingFieldErrorMessage = 'mangler'
    const validationErrorMessage = 'har en valideringsfeil'

    const resetForm = () => {
        setDisplayName('')
        setUniformName('')
        setOverrideUniformName(false)
        setUniformNameErrorMsg('')
        setSelectedSection(O.none())

        setSelectedAutonomyLevel(teamAutonomyLevels[0])
        setAdditionalInformation('')
        setSubmitButtonClicked(false)
    }

    const formErrors: FormError[] = useMemo(
        () =>
            pipe(
                [
                    { guard: displayName === '', field: displayNameLabel, errorMessage: missingFieldErrorMessage },
                    { guard: uniformName === '', field: uniformNameLabel, errorMessage: missingFieldErrorMessage },
                    {
                        guard: '' !== uniformNameErrorMsg,
                        field: uniformNameLabel,
                        errorMessage: validationErrorMessage,
                    },
                    { guard: O.isNone(selectedSection), field: sectionLabel, errorMessage: missingFieldErrorMessage },
                ],
                (errors) => A.zipWith(A.range(0, errors.length), errors, (idx, error) => ({ id: idx, ...error })),
                A.flatMap((mapping) =>
                    mapping.guard ? [{ id: mapping.id, field: mapping.field, errorMessage: mapping.errorMessage }] : []
                )
            ),
        [displayName, uniformName, selectedSection, uniformNameErrorMsg]
    )

    const [formSubmissionResult, setFormSubmissionResult] = useState<FormSubmissionResultProps>({
        loading: false,
        formSubmissionResult: O.none(),
    })

    useEffect(() => {
        Effect.gen(function* (_) {
            const sections: DisplaySSBSection[] = yield* _(
                Klass.fetchSSBSectionInformation().pipe(
                    Effect.map((sections: Klass.SSBSections) =>
                        sections.map((section) => ({ id: section.code, title: `${section.code} - ${section.name}` }))
                    )
                )
            )

            const storedUserProfile = localStorage.getItem('userProfile')
            const maybeUserProfile: O.Option<User> = O.fromNullable(storedUserProfile).pipe(O.map(JSON.parse))

            const userProfile: User = yield* _(
                Effect.try({
                    try: () => O.getOrThrow(maybeUserProfile),
                    catch: (error) => new Error(`Element not present: ${error}`),
                })
            )
            // Setting the selectedSection won't be visible beause of a ssb-component bug: https://github.com/statisticsnorway/ssb-component-library/pull/1111
            //const sectionCode = yield* getUserSectionCode(userProfile.principal_name)
            setUser(O.some(userProfile))
            setSections(sections)
            //setSelectedSection(A.findFirst(sections, (s) => s.id === sectionCode))
        }).pipe(Effect.runPromise)
    }, [])

    const handleSubmit = (event: Event) => {
        event.preventDefault()
        // Only submit the form if no form errors are present
        if (A.isEmptyArray(formErrors)) {
            const userPrincipalName = O.getOrThrow(user).principal_name
            const req: CreateTeamRequest = {
                teamDisplayName: displayName,
                uniformTeamName: uniformName,
                sectionCode: O.getOrThrow(selectedSection).id.toString(),
                additionalInformation: `This PR was created through Dapla Ctrl. Additional information from user ${userPrincipalName}:\n ${additionalInformation}`,
                autonomyLevel: selectedAutonomyLevel.id,
                features: [],
            }

            setFormSubmissionResult({ loading: true, formSubmissionResult: O.none() })

            Effect.gen(function* () {
                const clientResponse = yield* createTeam(req)
                yield* Console.log('ClientResponse', clientResponse)
                return O.some(
                    clientResponse.status !== 200
                        ? {
                              success: false,
                              message: `Det oppstod en feil ved opprettelse av team. Statuskode: ${clientResponse.status}`,
                          }
                        : { success: true, message: 'Opprettelse av team ble registert.' }
                )
            })
                .pipe(
                    Effect.catchTags({
                        ResponseError: (error) => Effect.succeed(O.some({ success: false, message: error.message })),
                        RequestError: (error) => Effect.succeed(O.some({ success: false, message: error.message })),
                        BodyError: (error) =>
                            Effect.succeed(
                                O.some({ success: false, message: `Failed to parse body: ${error.reason._tag}` })
                            ),
                    }),
                    Effect.runPromise
                )
                .then((res: O.Option<{ success: boolean; message: string }>) => {
                    if (
                        Utils.option(
                            res,
                            () => false,
                            (r) => r.success
                        )
                    ) {
                        resetForm()
                    }
                    setFormSubmissionResult({ loading: false, formSubmissionResult: res })
                })
        }
    }

    const toggleUniformNameInput = (checkboxes: string[]): void => {
        const isOverridden = checkboxes.includes('override')
        setOverrideUniformName(isOverridden)
        // If the user unselects the override option generate the uniform name
        // based on the display name again and clear all errors.
        if (!isOverridden) {
            setUniformNameErrorMsg('')
            setUniformName(generateUniformName(displayName))
        }
    }

    const generateUniformName = (displayName: string): string =>
        displayName
            .toLowerCase()
            .replaceAll('team ', '')
            .replaceAll('æ', 'ae')
            .replaceAll('ø', 'oe')
            .replaceAll('å', 'aa')
            .slice(0, uniformNameLengthLimit)
            .trim()
            .replaceAll(' ', '-')

    const validateUniformName = (name: string): O.Option<string> => {
        const validUniformName = generateUniformName(name)
        if (name.length > uniformNameLengthLimit) {
            return O.some(`Teknisk navn kan ikke være lengre enn ${uniformNameLengthLimit} tegn`)
        } else if (validUniformName !== name) {
            return O.some('Teknisk navn er ugyldig')
        } else {
            return O.none()
        }
    }

    const handleUniformNameInput = (name: string): void => {
        O.match(validateUniformName(name), {
            onNone: () => {
                setUniformNameErrorMsg('')
                setUniformName(name)
            },
            onSome: (errorMsg: string) => setUniformNameErrorMsg(errorMsg),
        })
    }

    const renderUniformNameField = () => {
        const label = (
            <C.Glossary className={styles.uniform_name_glossary} explanation={uniformNameGlossaryExplanation}>
                {uniformNameLabel}
            </C.Glossary>
        )
        return !overrideUniformName ? (
            <div className='ssb-input'>
                {label}
                <input className={styles.preview} type='text' value={uniformName} readOnly />
            </div>
        ) : (
            <C.Input
                label={label}
                id='uniform_name'
                type='text'
                value={uniformName}
                error={!!uniformNameErrorMsg}
                errorMessage={uniformNameErrorMsg}
                handleChange={handleUniformNameInput}
                maxLength={`${uniformNameLengthLimit}`}
            />
        )
    }

    const renderTeamOwnerCard = (isLoading: boolean) =>
        isLoading ? (
            <Skeleton variant='rectangular' animation='wave' width={350} height={200} />
        ) : (
            <Card className={styles.teamowner} title='Teamansvarlig (Managers)'>
                <Text>
                    {`${Utils.option(
                        user,
                        () => 'loading',
                        (u: User) => u.display_name
                    )} blir teamansvarlig for dette teamet. Hvis noen andre skal være ansvarlig kan det oppgis i feltet `}
                    <em>Tilleggsinformasjon</em>.
                </Text>
            </Card>
        )

    const renderContent = () => (
        <form className={styles.form} onSubmit={(e: React.FormEvent<HTMLFormElement>) => e.preventDefault()}>
            <Input
                label={<Glossary explanation={teamNameGlossaryExplanation}>{displayNameLabel}</Glossary>}
                id='display_name'
                type='text'
                handleChange={(value: string) => {
                    const uniformName = generateUniformName(value)
                    setUniformName(uniformName)
                    setDisplayName(value)
                }}
                value={displayName}
            />
            <CheckboxGroup
                onChange={(checkboxes: string[]) => toggleUniformNameInput(checkboxes)}
                orientation='column'
                items={[{ label: 'Overstyr teknisk navn?', value: 'override' }]}
            />
            {renderUniformNameField()}
            {/* NOTE: Changes to `selectedSection`, except from the `onSelect` function, doesn't re-render the component because the component is bugged. */}
            <Dropdown
                className={styles.section}
                header={<Glossary explanation={sectionGlossaryExplanation}>{sectionLabel}</Glossary>}
                searchable
                items={sections}
                selectedItem={O.getOrElse(selectedSection, () => undefined)}
                onSelect={(section: DisplaySSBSection) => setSelectedSection(O.some(section))}
            />
            <Dropdown
                header={<Glossary explanation={autonomyLevelGlossaryExplanation}>{'Autonomitetsnivå'}</Glossary>}
                selectedItem={selectedAutonomyLevel}
                items={teamAutonomyLevels}
                onSelect={(autonomyLevel: DisplayAutonomyLevel) => setSelectedAutonomyLevel(autonomyLevel)}
            />
            <TextArea
                label={
                    <Glossary explanation={additionalInformationGlossaryExplanation}>{'Tilleggsinformasjon'}</Glossary>
                }
                cols={40}
                rows={5}
                handleChange={setAdditionalInformation}
            />
            <Button
                className={styles.submitButton}
                type='submit'
                onClick={(e: Event) => {
                    setSubmitButtonClicked(true)
                    handleSubmit(e)
                }}
            >
                Opprett Team
            </Button>
            {renderTeamOwnerCard(O.isNone(user))}
            {submitButtonClicked && A.isNonEmptyArray(formErrors) && (
                <Dialog className={styles.warning} type='warning' title={'Valideringsfeil i skjemaet'}>
                    <div>
                        <p>{'Skjemaet har noen feil:'}</p>
                        <ul>
                            {formErrors.map((formError) => (
                                <li key={formError.id}>{`${formError.field} ${formError.errorMessage}`}</li>
                            ))}
                        </ul>
                    </div>
                </Dialog>
            )}
            <FormSubmissionResult {...formSubmissionResult} />
        </form>
    )

    return <PageLayout title='Opprett Team' content={renderContent()} />
}

export default CreateTeamForm
