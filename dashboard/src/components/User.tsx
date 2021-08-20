import React from 'react'
import { IUserSchema } from '@/schemas/user'
import { Avatar } from 'baseui/avatar'
import Text from './Text'

export interface IUserProps {
    user: IUserSchema
    size?: 'scale800' | 'scale1000' | 'scale1200' | 'scale1400' | '64px'
}

export default function User({ user, size = 'scale800' }: IUserProps) {
    return (
        <div
            style={{
                display: 'flex',
                alignItems: 'center',
                gap: 10,
            }}
        >
            <Avatar size={size} name={`${user.first_name} ${user.last_name}`} src={user.avatar_url} />
            <Text>
                {user.first_name} {user.last_name}
            </Text>
        </div>
    )
}