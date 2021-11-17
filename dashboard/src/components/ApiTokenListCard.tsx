import React, { useCallback, useState } from 'react'
import Card from '@/components/Card'
import { createApiToken } from '@/services/api_token'
import { usePage } from '@/hooks/usePage'
import { ICreateApiTokenSchema } from '@/schemas/api_token'
import ApiTokenForm from '@/components/ApiTokenForm'
import { formatDateTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { useFetchApiTokens } from '@/hooks/useFetchApiTokens'
import { resourceIconMapping } from '@/consts'
import { useStyletron } from 'baseui'
import { Input } from 'baseui/input'
import { TiClipboard } from 'react-icons/ti'
import { Notification } from 'baseui/notification'
import CopyToClipboard from 'react-copy-to-clipboard'

export default function ApiTokenListCard() {
    const [page] = usePage()
    const apiTokensInfo = useFetchApiTokens(page)
    const [theTokenWishToShow, setTheTokenWishToShow] = useState<string>()
    const [isCreateApiTokenOpen, setIsCreateApiTokenOpen] = useState(false)
    const [copyNotification, setCopyNotification] = useState<string>()
    const handleCreateApiToken = useCallback(
        async (data: ICreateApiTokenSchema) => {
            const apiToken = await createApiToken(data)
            setCopyNotification(undefined)
            setTheTokenWishToShow(apiToken.token)
            await apiTokensInfo.refetch()
            setIsCreateApiTokenOpen(false)
        },
        [apiTokensInfo]
    )
    const [t] = useTranslation()
    const [, theme] = useStyletron()

    return (
        <Card
            title={t('sth list', [t('api token')])}
            titleIcon={resourceIconMapping.api_token}
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateApiTokenOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <Table
                isLoading={apiTokensInfo.isLoading}
                columns={[
                    t('name'),
                    t('scopes'),
                    t('description'),
                    t('last_used_at'),
                    t('expired_at'),
                    t('created_at'),
                ]}
                data={
                    apiTokensInfo.data?.items.map((apiToken) => [
                        <Link key={apiToken.uid} to={`/api_tokens/${apiToken.name}`}>
                            {apiToken.name}
                        </Link>,
                        apiToken.scopes.join(', '),
                        apiToken.description,
                        apiToken.last_used_at ? formatDateTime(apiToken.last_used_at) : '-',
                        <span
                            key={apiToken.uid}
                            style={{
                                color: apiToken.is_expired ? theme.colors.negative : theme.colors.positive,
                            }}
                        >
                            {apiToken.expired_at ? formatDateTime(apiToken.expired_at) : '-'}
                        </span>,
                        formatDateTime(apiToken.created_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: apiTokensInfo.data?.start,
                    count: apiTokensInfo.data?.count,
                    total: apiTokensInfo.data?.total,
                    afterPageChange: () => {
                        apiTokensInfo.refetch()
                    },
                }}
            />
            <Modal
                isOpen={isCreateApiTokenOpen}
                onClose={() => setIsCreateApiTokenOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('create sth', [t('api token')])}</ModalHeader>
                <ModalBody>
                    <ApiTokenForm onSubmit={handleCreateApiToken} />
                </ModalBody>
            </Modal>
            <Modal
                isOpen={!!theTokenWishToShow}
                onClose={() => setTheTokenWishToShow(undefined)}
                size='default'
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('api token only show once time tips')}</ModalHeader>
                <ModalBody>
                    <div
                        style={{
                            display: 'flex',
                            gap: 10,
                        }}
                    >
                        <div
                            style={{
                                display: 'flex',
                                flexDirection: 'column',
                                gap: 4,
                                flexGrow: 1,
                            }}
                        >
                            <Input value={theTokenWishToShow} disabled />
                            {copyNotification && (
                                <Notification
                                    closeable
                                    onClose={() => setCopyNotification(undefined)}
                                    kind='positive'
                                    overrides={{
                                        Body: {
                                            style: {
                                                width: '100%',
                                                boxSizing: 'border-box',
                                                padding: '8px !important',
                                                borderRadius: '3px !important',
                                                fontSize: '13px !important',
                                            },
                                        },
                                    }}
                                >
                                    {copyNotification}
                                </Notification>
                            )}
                        </div>
                        <div>
                            <CopyToClipboard
                                text={theTokenWishToShow ?? ''}
                                onCopy={() => {
                                    setCopyNotification(t('copied to clipboard'))
                                }}
                            >
                                <Button startEnhancer={<TiClipboard size={14} />} kind='secondary'>
                                    {t('copy')}
                                </Button>
                            </CopyToClipboard>
                        </div>
                    </div>
                </ModalBody>
            </Modal>
        </Card>
    )
}
