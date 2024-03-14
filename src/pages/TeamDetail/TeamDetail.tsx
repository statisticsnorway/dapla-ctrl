/* eslint-disable react-hooks/exhaustive-deps */
import pageStyles from '../../components/PageLayout/pagelayout.module.scss'
import styles from './teamDetail.module.scss'

import { DropdownItems, TabProps } from '../../@types/pageTypes'

import { useCallback, useContext, useEffect, useState } from 'react'
import PageLayout from '../../components/PageLayout/PageLayout'
import { TeamDetailData, getTeamDetail, Team, SharedBuckets, addUserToGroups } from '../../services/teamDetail'
import { useParams } from 'react-router-dom'
import { ApiError, TokenData, fetchUserInformationFromAuthToken } from '../../utils/services'

import { DaplaCtrlContext } from '../../provider/DaplaCtrlProvider'
import Table, { TableData } from '../../components/Table/Table'
import { formatDisplayName, getGroupType } from '../../utils/utils'
import {
  Text,
  Dialog,
  LeadParagraph,
  Divider,
  Tabs,
  Button,
  Input,
  Dropdown,
  Tag,
} from '@statisticsnorway/ssb-component-library'
import PageSkeleton from '../../components/PageSkeleton/PageSkeleton'
import { Skeleton, CircularProgress } from '@mui/material'
import { XCircle } from 'react-feather'
import FormattedTableColumn from '../../components/FormattedTableColumn'
import SidebarModal from '../../components/SidebarModal/SidebarModal'

const TEAM_USERS_TAB = {
  title: 'Teammedlemmer',
  path: 'team',
  columns: [
    {
      id: 'navn',
      label: 'Navn',
    },
    {
      id: 'gruppe',
      label: 'Gruppe',
    },
    {
      id: 'epost',
      label: 'Epost ?',
    },
  ],
}

const SHARED_BUCKETS_TAB = {
  title: 'Delte data',
  path: 'sharedBuckets',
  columns: [
    {
      id: 'navn',
      label: 'Navn',
    },
    {
      id: 'tilgang',
      label: 'Tilgang',
    },
    { id: 'delte_data', label: 'Delte data' },
    { id: 'antall_personer', label: 'Antall personer' },
  ],
}

const defaultSelectedItem = {
  id: 'velg',
  title: 'Velg ...',
}

const TeamDetail = () => {
  const [activeTab, setActiveTab] = useState<TabProps | string>(TEAM_USERS_TAB)
  const [tokenData, setTokenData] = useState<TokenData>()

  const { setBreadcrumbTeamDetailDisplayName } = useContext(DaplaCtrlContext)
  const [error, setError] = useState<ApiError | undefined>()
  const [loadingTeamData, setLoadingTeamData] = useState<boolean>(true)
  const [teamDetailData, setTeamDetailData] = useState<TeamDetailData>()
  const [teamDetailTableTitle, setTeamDetailTableTitle] = useState<string>(TEAM_USERS_TAB.title)
  const [teamDetailTableHeaderColumns, setTeamDetailTableHeaderColumns] = useState<TableData['columns']>(
    TEAM_USERS_TAB.columns
  )
  const [teamDetailTableData, setTeamDetailTableData] = useState<TableData['data']>()

  const [openSidebar, setOpenSidebar] = useState<boolean>(false)
  const [email, setEmail] = useState({
    error: false,
    errorMessage: `Ugyldig epost`,
    value: '',
  })
  const [selectedItem, setSelectedItem] = useState(defaultSelectedItem)
  const [teamGroupTags, setTeamGroupTags] = useState<DropdownItems[]>([])
  const [teamGroupTagsError, setTeamGroupTagsError] = useState({
    error: false,
    errorMessage: 'Velg minst én tilgangsgruppe',
  })
  const [addUserToTeamErrors, setAddUserToTeamErrors] = useState<Array<string>>([])
  const [showSpinner, setShowSpinner] = useState<boolean>(false)

  const { teamId } = useParams<{ teamId: string }>()
  const teamDetailTab = (activeTab as TabProps)?.path ?? activeTab

  const prepTeamData = useCallback(
    (response: TeamDetailData): TableData['data'] => {
      const sharedBucketsTab = SHARED_BUCKETS_TAB.path
      if (teamDetailTab === sharedBucketsTab) {
        const sharedBuckets = (response[sharedBucketsTab] as SharedBuckets).items
        if (!sharedBuckets) return []

        return sharedBuckets.map(({ short_name, bucket_name, metrics }) => {
          const teams_count = metrics?.teams_count
          return {
            id: short_name,
            navn: <FormattedTableColumn href={`/${teamId}/${short_name}`} linkText={short_name} text={bucket_name} />,
            tilgang: typeof teams_count === 'number' ? `${teams_count} team` : teams_count,
            delte_data: metrics?.groups_count,
            antall_personer: metrics?.users_count,
          }
        })
      } else {
        const teamUsers = (response[TEAM_USERS_TAB.path] as Team).users
        if (!teamUsers) return []

        return teamUsers.map(({ display_name, principal_name, section_name, groups }) => {
          return {
            id: formatDisplayName(display_name),
            navn: (
              <FormattedTableColumn
                href={`/teammedlemmer/${principal_name}`}
                linkText={formatDisplayName(display_name)}
                text={section_name}
              />
            ),
            seksjon: section_name, // Makes section name searchable and sortable in table by including the field
            gruppe: groups
              ?.filter((group) => group.uniform_name.startsWith((response.team as Team).uniform_name))
              .map((group) => getGroupType(group.uniform_name))
              .join(', '),
            epost: principal_name,
          }
        })
      }
    },
    [activeTab]
  )

  useEffect(() => {
    if (!teamId) return
    fetchUserInformationFromAuthToken()
      .then((tokenData) => setTokenData(tokenData))
      .catch((error) => setError(error as ApiError))
    getTeamDetail(teamId)
      .then((response) => {
        const formattedResponse = response as TeamDetailData
        setTeamDetailData(formattedResponse)
        setTeamDetailTableData(prepTeamData(formattedResponse))

        const displayName = formatDisplayName((formattedResponse.team as Team).display_name)
        setBreadcrumbTeamDetailDisplayName({ displayName })
      })
      .finally(() => setLoadingTeamData(false))
      .catch((error) => {
        setError(error as ApiError)
      })
  }, [])

  useEffect(() => {
    if (teamDetailData) {
      if (teamDetailTab === SHARED_BUCKETS_TAB.path) {
        setTeamDetailTableTitle(SHARED_BUCKETS_TAB.title)
        setTeamDetailTableHeaderColumns(SHARED_BUCKETS_TAB.columns)
      } else {
        setTeamDetailTableTitle(TEAM_USERS_TAB.title)
        setTeamDetailTableHeaderColumns(TEAM_USERS_TAB.columns)
      }
      setTeamDetailTableData(prepTeamData(teamDetailData))
    }
  }, [prepTeamData])

  const handleTabClick = (tab: string) => setActiveTab(tab)

  const renderErrorAlert = () => {
    return (
      <Dialog type='warning' title='Could not fetch teams'>
        {`${error?.code} - ${error?.message}`}
      </Dialog>
    )
  }

  const renderContent = () => {
    if (error) return renderErrorAlert()
    if (loadingTeamData) return <PageSkeleton hasDescription />

    if (teamDetailData && teamDetailTableHeaderColumns && teamDetailTableData) {
      return (
        <>
          <LeadParagraph className={pageStyles.description}>
            <Text medium className={pageStyles.descriptionSpacing}>
              {(teamDetailData.team as Team).uniform_name ?? ''}
            </Text>
            <Text medium>{formatDisplayName((teamDetailData.team as Team).manager?.display_name ?? '')}</Text>
            <Text medium>{(teamDetailData.team as Team).section_name ?? ''}</Text>
          </LeadParagraph>
          <Tabs
            onClick={handleTabClick}
            activeOnInit={TEAM_USERS_TAB.path}
            items={[
              {
                title: `${TEAM_USERS_TAB.title} (${(teamDetailData?.team as Team).users?.length ?? 0})`,
                path: TEAM_USERS_TAB.path,
              },
              {
                title: `${SHARED_BUCKETS_TAB.title} (${(teamDetailData?.sharedBuckets as SharedBuckets).items?.length ?? 0})`,
                path: SHARED_BUCKETS_TAB.path,
              },
            ]}
          />
          <Divider dark />
          <Table
            title={teamDetailTableTitle}
            columns={teamDetailTableHeaderColumns}
            data={teamDetailTableData as TableData['data']}
          />
        </>
      )
    }
  }

  const handleAddTeamGroupTag = (item: DropdownItems) => {
    const teamGroupsTags = [...teamGroupTags, item].reduce((acc: DropdownItems[], dropdownItem: DropdownItems) => {
      const ids = acc.map((obj) => obj.id)
      if (!ids.includes(dropdownItem.id)) {
        acc.push(dropdownItem)
      }
      return acc
    }, [])
    setTeamGroupTags(teamGroupsTags)
    setTeamGroupTagsError({ ...teamGroupTagsError, error: false })
  }

  const handleDeleteGroupTag = (item: DropdownItems) => {
    const teamGroupsTags = teamGroupTags.filter((items) => items !== item)
    setTeamGroupTags(teamGroupsTags)
  }

  const isUserInputValid = (value?: string) => {
    const regEx = /^[\w-]+@ssb\.no$/
    const userVal = value || email.value
    const testUser = userVal.match(regEx)
    return !!testUser
  }

  const handleAddUserOnSubmit = () => {
    if (email.value === '') setEmail({ ...email, error: true })
    if (!teamGroupTags.length)
      setTeamGroupTagsError({
        ...teamGroupTagsError,
        error: true,
      })

    if (email.value !== '' && teamGroupTags.length) {
      setAddUserToTeamErrors([])
      setShowSpinner(true)
      addUserToGroups(
        teamGroupTags.map((group) => group.id),
        email.value
      )
        .then((response) => {
          const errorsList = response
            .map(({ status, detail }) => {
              if ((detail && status === 'ERROR') || (detail && status === 'IGNORED')) {
                return detail
              }
              return ''
            })
            .filter((str) => str !== '')

          if (!errorsList.length) {
            setAddUserToTeamErrors(errorsList)
          } else {
            setOpenSidebar(false)
            // setEmail({ ...email, value: '' })
            setSelectedItem({ ...defaultSelectedItem })
            // setTeamGroupTags([]) // TODO: Re-implement when clearing input fields work
          }
        })
        .catch((e) => setAddUserToTeamErrors(e.message))
        .finally(() => setShowSpinner(false))
    }
  }

  const renderSidebarModalAlert = () => {
    return (
      <div className={styles.modalBodyDialog}>
        <Dialog type='info'>Det kan ta opp til 45 minutter før personen kan bruke tilgangen</Dialog>
        {addUserToTeamErrors.length ? (
          <Dialog type='warning'>
            {typeof addUserToTeamErrors === 'string' ? (
              addUserToTeamErrors
            ) : (
              <ul>
                {addUserToTeamErrors.map((errors) => (
                  <li>{errors}</li>
                ))}
              </ul>
            )}
          </Dialog>
        ) : null}
        {showSpinner && <CircularProgress />}
      </div>
    )
  }

  const renderSidebarModal = () => {
    if (teamDetailData) {
      const teamGroups = (teamDetailData?.team as Team).groups ?? []
      return (
        <SidebarModal
          open={openSidebar}
          onClose={() => setOpenSidebar(false)}
          header={{
            modalType: 'Medlem',
            modalTitle: `${(teamDetailData?.team as Team).display_name}`,
            modalDescription: `${(teamDetailData?.team as Team).uniform_name}`,
          }}
          footer={{
            submitButtonText: 'Legg til medlem',
            handleSubmit: handleAddUserOnSubmit,
          }}
          body={{
            modalBodyTitle: 'Legg person til teamet',
            modalBody: (
              <>
                <Input
                  className={styles.inputSpacing}
                  label='Kort epost'
                  value={email.value}
                  error={email.error}
                  errorMessage={email.errorMessage}
                  onBlur={() =>
                    setEmail({
                      ...email,
                      error: !isUserInputValid(),
                    })
                  }
                  handleChange={(value: string) =>
                    setEmail({
                      ...email,
                      value,
                      error: email.error ? !isUserInputValid(value) : false,
                    })
                  }
                />
                <Dropdown
                  className={styles.dropdownSpacing}
                  header='Tilgangsgrupper(r)'
                  selectedItem={selectedItem}
                  items={teamGroups.map(({ uniform_name }) => ({
                    id: uniform_name,
                    title: getGroupType(uniform_name),
                  }))}
                  onSelect={handleAddTeamGroupTag}
                  error={teamGroupTagsError.error}
                  errorMessage={teamGroupTagsError.errorMessage}
                />
                <div className={styles.tagsContainer}>
                  {teamGroupTags &&
                    teamGroupTags.map((group) => (
                      <Tag icon={<XCircle size={14} />} onClick={() => handleDeleteGroupTag(group)}>
                        {group.title}
                      </Tag>
                    ))}
                </div>
                {renderSidebarModalAlert()}
              </>
            ),
          }}
        />
      )
    }
  }

  const teamManager = teamDetailData ? (teamDetailData?.team as Team).manager?.principal_name : ''
  return (
    <>
      {renderSidebarModal()}
      <PageLayout
        title={
          !loadingTeamData && teamDetailData ? (
            (teamDetailData.team as Team).display_name
          ) : (
            <Skeleton variant='rectangular' animation='wave' width={350} height={90} />
          )
        }
        content={renderContent()}
        button={
          tokenData?.email === teamManager ? (
            <Button onClick={() => setOpenSidebar(true)}>+ Nytt medlem</Button>
          ) : undefined
        }
      />
    </>
  )
}

export default TeamDetail
