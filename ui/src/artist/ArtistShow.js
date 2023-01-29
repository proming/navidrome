import React, { useState, createElement, useEffect } from 'react'
import { useMediaQuery } from '@material-ui/core'
import {
  useShowController,
  ShowContextProvider,
  useRecordContext,
  useShowContext,
  Pagination,
  ReferenceManyField,
} from 'react-admin'
import { useAlbumsPerPage } from '../common'
import subsonic from '../subsonic'
import AlbumGridView from '../album/AlbumGridView'
import MobileArtistDetails from './MobileArtistDetails'
import DesktopArtistDetails from './DesktopArtistDetails'
import { AddToPlaylistDialog } from '../dialogs'
import AlbumInfo from '../album/AlbumInfo'
import ExpandInfoDialog from '../dialogs/ExpandInfoDialog'

const ArtistDetails = (props) => {
  const record = useRecordContext(props)
  const isDesktop = useMediaQuery((theme) => theme.breakpoints.up('sm'))
  const [artistInfo, setArtistInfo] = useState()

  const biography =
    artistInfo?.biography?.replace(new RegExp('<.*>', 'g'), '') ||
    record.biography
  const img = artistInfo?.largeImageUrl || record.largeImageUrl

  useEffect(() => {
    subsonic
      .getArtistInfo(record.id)
      .then((resp) => resp.json['subsonic-response'])
      .then((data) => {
        if (data.status === 'ok') {
          setArtistInfo(data.artistInfo)
        }
      })
      .catch((e) => {
        console.error('error on artist page', e)
      })
  }, [record.id])

  const component = isDesktop ? DesktopArtistDetails : MobileArtistDetails
  return (
    <>
      {createElement(component, {
        img,
        artistInfo,
        record,
        biography,
      })}
    </>
  )
}

const AlbumShowLayout = (props) => {
  const { width } = props
  const showContext = useShowContext(props)
  const record = useRecordContext()
  const [perPage, perPageOptions] = useAlbumsPerPage(width)

  return (
    <>
      {record && <ArtistDetails />}
      {record && (
        <>
          <ReferenceManyField
            {...showContext}
            addLabel={false}
            reference="album"
            target="artist_id"
            sort={{ field: 'max_year', order: 'ASC' }}
            filter={{ artist_id: record?.id }}
            perPage={perPage}
            pagination={<Pagination rowsPerPageOptions={perPageOptions} />}
          >
            <AlbumGridView {...props} />
          </ReferenceManyField>
          <AddToPlaylistDialog />
          <ExpandInfoDialog content={<AlbumInfo />} />
        </>
      )}
    </>
  )
}

const ArtistShow = (props) => {
  const controllerProps = useShowController(props)
  return (
    <ShowContextProvider value={controllerProps}>
      <AlbumShowLayout {...controllerProps} />
    </ShowContextProvider>
  )
}

export default ArtistShow
